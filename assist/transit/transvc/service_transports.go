package transvc

import (
	"context"
	"sort"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/assist/transit/transutil"
	"go.stevenxie.me/api/v2/location"
)

func (svc *service) NearbyTransports(
	ctx context.Context,
	pos location.Coordinates,
	opts ...transit.NearbyTransportsOption,
) ([]transit.Transport, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc((*service).NearbyTransports),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*service).NearbyTransports),
		"position":        pos,
	}).WithContext(ctx)

	var opt transit.NearbyTransportsOptions
	for _, apply := range opts {
		apply(&opt)
	}
	{
		fields := make(logrus.Fields)
		if r := opt.Radius; r > 0 {
			fields["radius"] = r
		}
		if l := opt.Limit; l > 0 {
			fields["limit"] = l
		}
		if m := opt.MaxStations; m > 0 {
			fields["max_stations"] = m
		}
		log = log.WithFields(fields)
	}

	log.Trace("Finding departures...")
	nds, err := svc.loc.FindDepartures(
		ctx,
		pos,
		func(findOpt *transit.FindDeparturesOptions) {
			findOpt.MaxPerTransport = 1
			if r := opt.Radius; r > 0 {
				findOpt.Radius = r
			}
			if m := opt.MaxStations; m > 0 {
				findOpt.MaxStations = m
			}
		},
	)
	if err != nil {
		log.WithError(err).Trace("Failed to get nearby departures.")
		return nil, err
	}

	// Build unique map of Transports and their station distances.
	var (
		tpm = make(map[uint32]*transit.Transport) // transports map
		ddm = make(map[uint32]int)                // departure distances map
	)
	for i := range nds {
		tp := nds[i].Transport
		hash := transutil.HashTransport(tp)
		if _, ok := tpm[hash]; !ok {
			tpm[hash] = tp
			ddm[hash] = nds[i].Distance
		}
	}
	log.WithFields(logrus.Fields{
		"transports_map": tpm,
		"distances_map":  ddm,
	}).Trace("Built uniqueness maps.")

	// Build and sort hash list from map.
	hashes := make([]uint32, 0, len(ddm))
	for h := range ddm {
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		return ddm[hashes[i]] < ddm[hashes[j]]
	})
	log.WithField("hashes", hashes).Trace("Built sorted hash list.")

	// Construct transports list.
	tps := make([]transit.Transport, len(hashes))
	for i, h := range hashes {
		tps[i] = *tpm[h]
	}
	log.WithField("transports", tps).Trace("Built sorted transports list.")

	// Apply limit.
	if l := opt.Limit; len(nds) > l {
		nds = nds[:l]
		log.
			WithField("transports", tps).
			Trace("Applied transports limit.")
	}

	return tps, nil
}
