package transvc

import (
	"context"
	"sort"

	"github.com/openlyinc/pointy"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/transutil"
	"go.stevenxie.me/api/location"
)

func (svc service) NearbyTransports(
	ctx context.Context,
	coords location.Coordinates,
	opts ...transit.NearbyTransportsOption,
) ([]transit.Transport, error) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.NearbyTransports),
		"coordinates":     coords,
	}).WithContext(ctx)

	var cfg transit.NearbyTransportsConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	{
		fields := make(logrus.Fields)
		if r := cfg.Radius; r != nil {
			fields["radius"] = *r
		}
		if l := cfg.Limit; l != nil {
			fields["limit"] = *l
		}
		if m := cfg.MaxStations; m != nil {
			fields["max_stations"] = *m
		}
		log = log.WithFields(fields)
	}

	log.Trace("Getting nearby departures...")
	nds, err := svc.loc.NearbyDepartures(
		ctx,
		coords,
		func(ndCfg *transit.NearbyDeparturesConfig) {
			ndCfg.MaxPerTransport = pointy.Int(1)
			if r := cfg.Radius; r != nil {
				ndCfg.Radius = r
			}
			if m := cfg.MaxStations; m != nil {
				ndCfg.MaxStations = m
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
	if l := cfg.Limit; (l != nil) && (len(nds) > *l) {
		nds = nds[:*l]
		log.WithFields(logrus.Fields{
			"limit":      l,
			"transports": tps,
		}).Trace("Applied transports limit.")
	}

	return tps, nil
}
