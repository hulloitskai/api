package transvc

import (
	"context"

	"go.stevenxie.me/gopkg/name"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewLocatorService creates a new transit.LocatorService.
func NewLocatorService(
	loc transit.Locator,
	opts ...svcutil.BasicOption,
) transit.LocatorService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return locatorService{
		loc: loc,
		log: logutil.AddComponent(cfg.Logger, (*locatorService)(nil)),
	}
}

type locatorService struct {
	loc transit.Locator
	log *logrus.Entry
}

var _ transit.LocatorService = (*locatorService)(nil)

func (svc locatorService) NearbyDepartures(
	ctx context.Context,
	pos location.Coordinates,
	opts ...transit.NearbyDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(svc.NearbyDepartures),
		"position":        pos,
	})

	// Derive config, add log fields.
	var cfg transit.NearbyDeparturesConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	{
		fields := make(logrus.Fields)
		if r := cfg.Radius; r != nil {
			fields["radius"] = r
		}
		if m := cfg.MaxStations; m != nil {
			fields["max_stations"] = *m
		}
		if m := cfg.MaxPerStation; m != nil {
			fields["max_per_station"] = *m
		}
		if m := cfg.MaxPerTransport; m != nil {
			fields["max_per_transport"] = *m
		}
		log = logrus.WithFields(fields)
	}

	log.Trace("Getting nearby departures...")
	nds, err := svc.loc.NearbyDepartures(ctx, pos, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to get nearby departures.")
		return nil, err
	}

	return nds, nil
}
