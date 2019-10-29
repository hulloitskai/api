package transvc

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewLocatorService creates a new transit.LocatorService.
func NewLocatorService(
	loc transit.Locator,
	opts ...basic.Option,
) transit.LocatorService {
	opt := basic.BuildOptions(opts...)
	return locatorService{
		loc:    loc,
		log:    logutil.WithComponent(opt.Logger, (*locatorService)(nil)),
		tracer: opt.Tracer,
	}
}

type locatorService struct {
	loc    transit.Locator
	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ transit.LocatorService = (*locatorService)(nil)

func (svc locatorService) NearbyDepartures(
	ctx context.Context,
	pos location.Coordinates,
	opts ...transit.NearbyDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(locatorService.NearbyDepartures),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(svc.NearbyDepartures),
		"position":        pos,
	})

	// Derive config, add log fields.
	var opt transit.NearbyDeparturesOptions
	for _, apply := range opts {
		apply(&opt)
	}
	{
		fields := make(logrus.Fields)
		if r := opt.Radius; r > 0 {
			fields["radius"] = r
		}
		if m := opt.MaxStations; m > 0 {
			fields["max_stations"] = m
		}
		if m := opt.MaxPerStation; m > 0 {
			fields["max_per_station"] = m
		}
		if m := opt.MaxPerTransport; m > 0 {
			fields["max_per_transport"] = m
		}
		log = logrus.WithFields(fields)
	}

	log.Trace("Getting nearby departures...")
	nds, err := svc.loc.NearbyDepartures(ctx, pos, opt)
	if err != nil {
		log.WithError(err).Error("Failed to get nearby departures.")
		return nil, err
	}

	return nds, nil
}
