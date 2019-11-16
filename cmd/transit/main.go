package main

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/k0kubun/pp"
	funk "github.com/thoas/go-funk"

	"go.stevenxie.me/gopkg/cmdutil"
	"go.stevenxie.me/gopkg/configutil"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/assist/transit/grt"
	"go.stevenxie.me/api/v2/assist/transit/heretrans"
	"go.stevenxie.me/api/v2/assist/transit/transvc"

	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/pkg/basic"
	"go.stevenxie.me/api/v2/pkg/here"
)

const _appID = "NJBvN0hkaR2Fv7bNpNpU"

func main() {
	if err := configutil.LoadEnv(); err != nil {
		cmdutil.Fatalf("Failed to load dotenv file: %v\n", err)
	}
	log := cmdutil.NewLogger()

	if len(os.Args) != 3 ||
		funk.ContainsString([]string{"-h", "--help"}, os.Args[1]) {
		cmdutil.Fatalf("Usage: %s <lat,lon> <route>\n", os.Args[0])
	}

	// Derive coordinates from args.
	var pos location.Coordinates
	{
		fields := strings.Split(os.Args[1], ",")
		if len(fields) != 2 {
			cmdutil.Fatalln("Invalid position argument format (should be <lat,lon>).")
		}
		convs := []struct {
			Dst *float64
			Src string
		}{
			{Dst: &pos.X, Src: fields[1]},
			{Dst: &pos.Y, Src: fields[0]},
		}
		var err error
		for _, c := range convs {
			if *c.Dst, err = strconv.ParseFloat(c.Src, 64); err != nil {
				cmdutil.Fatalf(
					"Failed to parse position component '%s' as float.\n",
					c.Src,
				)
			}
		}
	}

	var locsvc transit.LocatorService
	{
		client, err := here.NewClient(_appID)
		if err != nil {
			log.WithError(err).Fatal("Failed to create Here client.")
		}
		loc := heretrans.NewLocator(client)
		locsvc = transvc.NewLocatorService(loc, basic.WithLogger(log))
	}

	// Create service.
	var svc transit.Service
	{
		grt, err := grt.NewRealtimeSource(grt.WithLogger(log))
		if err != nil {
			log.WithError(err).Fatal("Failed to create RealTimeSource.")
		}
		svc = transvc.NewService(
			locsvc,
			transvc.WithLogger(log),
			transvc.WithRealtimeSource(grt, transit.OpCodeGRT),
		)
	}

	deps, err := svc.NearbyDepartures(
		context.Background(),
		pos,
		os.Args[2],
		transit.FindWithFuzzyMatch(true),
		transit.FindWithGroupByStation(true),
		transit.FindWithLimit(2),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to locate nearby departures.")
	}
	pp.Println(deps)
}
