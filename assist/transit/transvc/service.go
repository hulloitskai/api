package transvc

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.stevenxie.me/api/assist/assistutil"

	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
	"go.stevenxie.me/gopkg/zero"

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

	// Validate inputs.
	if route == "" {
		return nil, errors.New("transvc: route is empty")
	}

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

	// Filter based on operator, if applicable.
	if code := cfg.OperatorCode; code != "" {
		var filtered []transit.NearbyDeparture
		for i := range nds {
			if nds[i].Transport.Operator.Code == code {
				filtered = append(filtered, nds[i])
			}
		}
		nds = filtered
	}

	// If fuzzy-matching, update route to the closest matching route.
	if cfg.FuzzyMatch {
		// Basic input normalization.
		query := strings.ToLower(route)
		query = strings.TrimSpace(query)
		query = strings.TrimPrefix(query, "the ")
		query = strings.Trim(query, ".!?")
		query = assistutil.ReplaceNumberWords(query)

		// Construct search strings to match against.
		var (
			routes            = make([]string, len(nds))
			routesWithContext = make([]string, len(nds))
		)
		for i := range nds {
			tp := nds[i].Transport
			routes[i] = tp.Route
			routesWithContext[i] = assistutil.ReplaceNumberWords(fmt.Sprintf(
				"%s %s to %s",
				tp.Route, tp.Operator.Name, tp.Direction,
			))
		}

		matches := fuzzy.RankFind(query, routesWithContext)
		if len(matches) == 0 {
			return nil, errors.WithHintf(
				errors.New("transvc: no matching route"),
				"No nearby routes matching '%s'.", route,
			)
		}
		sort.Sort(matches)
		route = routes[matches[0].OriginalIndex]
	}

	// Filter based on route.
	{
		var filtered []transit.NearbyDeparture
		for i := range nds {
			if code := cfg.OperatorCode; code != "" {
				if nds[i].Transport.Operator.Code != code {
					continue
				}
			}
			if route == nds[i].Transport.Route {
				filtered = append(filtered, nds[i])
			}
		}
		if len(filtered) == 0 {
			return nil, errors.WithHintf(
				errors.New("transvc: no departures matching route"),
				"No departures found for '%s'.", route,
			)
		}
		nds = filtered
	}

	// Group by station, if enabled.
	if cfg.GroupByStation {
		var (
			included = make(map[string]zero.Struct)
			sorted   = make([]transit.NearbyDeparture, 0, len(nds))
		)
		for i := range nds {
			var (
				stn = nds[i].Station
				sid = stn.ID
			)
			if _, ok := included[sid]; !ok {
				sorted = append(sorted, nds[i])
				included[sid] = zero.Empty()

				// Find other matching stations by name.
				for j := i + 1; j < len(nds); j++ {
					otherStn := nds[j].Station
					if otherStn.Name == stn.Name {
						sorted = append(sorted, nds[j])
						included[otherStn.ID] = zero.Empty()
					}
				}
			}
		}
		nds = sorted
	}

	// Apply limit if it exists.
	if l := cfg.Limit; l > 0 {
		nds = nds[:l]
	}

	// Update with realtime departures times, if available.
	for i := range nds {
		times, err := svc.rts.GetDepartureTimes(
			ctx,
			*nds[i].Transport, *nds[i].Station,
		)
		if err != nil {
			if errors.Is(err, transit.ErrOperatorNotSupported) {
				continue
			}
			log.WithError(err).Error("Failed to get realtime departure times.")
			return nil, errors.Wrap(err, "transvc: get realtime departure times")
		}

		nd := &nds[i]
		nd.Times = times
		nd.Realtime = true
	}
	return nds, nil
}
