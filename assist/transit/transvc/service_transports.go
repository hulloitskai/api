package transvc

import (
	"context"

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
		return nil, err
	}

	// Build unique map of Transports.
	uniq := make(map[uint32]*transit.Transport)
	for i := range nds {
		tp := nds[i].Transport
		hash := transutil.HashTransport(tp)
		if _, ok := uniq[hash]; !ok {
			uniq[hash] = tp
		}
	}

	// Build list from uniq.
	tps := make([]transit.Transport, 0, len(uniq))
	for _, tp := range uniq {
		tps = append(tps, *tp)
	}

	// Apply limit.
	if l := cfg.Limit; (l != nil) && (len(nds) > *l) {
		nds = nds[:*l]
	}

	return tps, nil
}
