package transvc

import (
	"context"

	"github.com/schollz/closestmatch"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/svcutil"
)

// NewService creates a new transit.Service.
func NewService(
	loc transit.Locator,
	rts transit.RealtimeSource,
	opts ...svcutil.BasicOption,
) transit.Service {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		loc: loc,
		rts: rts,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	loc transit.Locator
	rts transit.RealtimeSource
	log *logrus.Entry
}

var _ transit.Service = (*service)(nil)

func (svc service) FindDepartures(
	ctx context.Context,
	route string,
	pos location.Coordinates,
	opts ...transit.FindDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	log := svc.log.WithFields(logrus.Fields{
		"method": name.OfMethod(service.FindDepartures),
		"route":  route,
		"pos":    pos,
	}).WithContext(ctx)

	var cfg transit.FindDeparturesConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.OperatorCode != "" {
		log = log.WithField("operator_code", cfg.OperatorCode)
	}

	nds, err := svc.loc.NearbyDepartures(
		ctx,
		pos,
		func(ndCfg *transit.NearbyDeparturesConfig) {
			if cfg.Radius != nil {
				ndCfg.Radius = cfg.Radius
			}
			if max := cfg.MaxStations; max != nil {
				ndCfg.MaxStations = max
			}
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to get nearby departures.")
		return nil, errors.Wrap(err, "transvc: get nearby departures")
	}

	// If fuzzy-matching, update route to the closest matching route.
	if cfg.FuzzyMatch {
		routes := make([]string, len(nds))
		for i := range nds {
			routes[i] = nds[i].Transport.Route
		}
		route = closestmatch.New(routes, []int{1}).Closest(route)
	}

	// Filter based on route.
	var ndsFiltered []transit.NearbyDeparture
	for i := range nds {
		if code := cfg.OperatorCode; code != "" {
			if nds[i].Transport.Operator.Code != code {
				continue
			}
		}
		if route == nds[i].Transport.Route {
			ndsFiltered = append(ndsFiltered, nds[i])
			if l := cfg.Limit; (l > 0) && (len(ndsFiltered) == l) {
				break
			}
		}
	}
	if len(ndsFiltered) == 0 {
		return nil, errors.WithHintf(
			errors.New("transvc: no departures matching route"),
			"No departures found for '%s'.", route,
		)
	}
	nds = nil // free nds

	// Update with realtime departures times, if available.
	for i, nd := range ndsFiltered {
		times, err := svc.rts.GetDepartureTimes(ctx, *nd.Transport, *nd.Station)
		if err != nil && !errors.Is(err, transit.ErrOperatorNotSupported) {
			log.WithError(err).Error("Failed to get realtime departure times.")
			return nil, errors.Wrap(err, "transvc: get realtime departure times")
		}

		nd := &ndsFiltered[i]
		nd.Times = times
		nd.Realtime = true
	}
	return ndsFiltered, nil
}
