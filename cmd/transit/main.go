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

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/grt"
	"go.stevenxie.me/api/assist/transit/heretrans"
	"go.stevenxie.me/api/assist/transit/transvc"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/here"
	"go.stevenxie.me/api/pkg/svcutil"
)

const _appID = "NJBvN0hkaR2Fv7bNpNpU"

func main() {
	if err := configutil.LoadEnv(); err != nil {
		cmdutil.Fatalf("Failed to load dotenv file: %v\n", err)
	}
	log := buildLogger()

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
				cmdutil.Fatalf("Parsing position component '%s' as float.\n", c.Src)
			}
		}
	}

	var loc transit.Locator
	{
		client, err := here.NewClient(_appID)
		if err != nil {
			log.WithError(err).Fatal("Creating Here client.")
		}
		loc = heretrans.NewLocator(client)
	}

	var rts transit.RealtimeSource
	{
		rts = grt.NewRealtimeSource(nil)
	}

	svc := transvc.NewService(loc, rts, svcutil.WithLogger(log))
	deps, err := svc.FindDepartures(
		context.Background(),
		os.Args[2],
		pos,
		transit.FindWithFuzzyMatch(),
		transit.FindWithLimit(2),
	)
	if err != nil {
		log.WithError(err).Fatal("Failed to locate nearby departures.")
	}
	pp.Println(deps)
}
