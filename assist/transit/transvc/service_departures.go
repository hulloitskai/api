package transvc

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/name"
	"go.stevenxie.me/gopkg/zero"

	"go.stevenxie.me/api/assist/assistutil"
	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/location"
)

func (svc service) FindDepartures(
	ctx context.Context,
	routeQuery string,
	coords location.Coordinates,
	opts ...transit.FindDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	log := svc.log.WithFields(logrus.Fields{
		"method":      name.OfMethod(service.FindDepartures),
		"route_query": routeQuery,
		"coordinates": coords,
	}).WithContext(ctx)

	// Validate inputs.
	if routeQuery == "" {
		return nil, errors.New("transvc: route is empty")
	}

	cfg := transit.FindDeparturesConfig{
		PreferRealtime: true,
		MaxTimes:       3,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if cfg.OperatorCode != "" {
		log = log.WithField("operator_code", cfg.OperatorCode)
	}

	nds, err := svc.loc.NearbyDepartures(
		ctx,
		coords,
		func(ndCfg *transit.NearbyDeparturesConfig) {
			ndCfg.MaxPerTransport = &cfg.MaxTimes
			if cfg.Radius != nil {
				ndCfg.Radius = cfg.Radius
			}
			if max := cfg.MaxStations; max != nil {
				ndCfg.MaxStations = max
			}
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "transvc: get nearby departures")
	}

	// Filter based on operator, if applicable.
	if code := cfg.OperatorCode; code != "" {
		log.WithField("code", code).Trace("Filtering by code...")
		var filtered []transit.NearbyDeparture
		for i := range nds {
			if nds[i].Transport.Operator.Code == code {
				filtered = append(filtered, nds[i])
			}
		}
		nds = filtered
	}

	// If fuzzy-matching, update route to the closest matching route.
	var route string
	if cfg.FuzzyMatch {
		// Input normalization.
		route = strings.ToLower(routeQuery)
		route = strings.TrimSpace(route)
		route = strings.TrimPrefix(route, "the ")
		route = strings.TrimSuffix(route, "th")
		route = strings.ReplaceAll(route, "be", "b")
		route = strings.Trim(route, ".!?")
		route = assistutil.ReplaceNumberWords(route)

		// Construct search strings to match against.
		var (
			routes            = make([]string, len(nds))
			routesWithContext = make([]string, len(nds))
		)
		for i := range nds {
			tp := nds[i].Transport
			routes[i] = tp.Route

			r := normalizeAdjacentNumAlpha(tp.Route)
			rwc := fmt.Sprintf("%s %s", tp.Operator.Name, r)
			rwc = assistutil.ReplaceNumberWords(rwc)
			routesWithContext[i] = rwc
		}

		log.
			WithField("targets", routesWithContext).
			Trace("Performing fuzzy search against targets.")
		matches := fuzzy.RankFindFold(route, routesWithContext)
		if len(matches) == 0 {
			return nil, errors.WithDetailf(
				errors.New("transvc: no matching route"),
				"No nearby routes matching '%s'.", routeQuery,
			)
		}
		sort.Sort(matches)
		route = routes[matches[0].OriginalIndex]
	} else {
		route = routeQuery
	}
	log = log.WithField("route", route)
	log.Trace("Derived route.")

	// Filter based on route.
	{
		log.Trace("Filtering results by route...")
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
			return nil, errors.WithDetailf(
				errors.New("transvc: no departures matching route"),
				"No departures found for '%s'.", routeQuery,
			)
		}
		nds = filtered
	}

	// Group by station, if enabled.
	if cfg.GroupByStation {
		log.Trace("Grouping results by station...")
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
	if l := cfg.Limit; (l > 0) && (len(nds) > l) {
		nds = nds[:l]
	}

	// Update with realtime departures times, if available.
	if cfg.PreferRealtime {
		log.Trace("Update results with realtime departures...")
		for i := range nds {
			if nds[i].Realtime {
				continue // departure already is realtime
			}

			var (
				tp  = nds[i].Transport
				stn = nds[i].Station
			)
			log := log.WithFields(logrus.Fields{
				"direction": tp.Direction,
				"station":   stn.Name,
			})
			log.Trace("Getting realtime departure...")

			times, err := svc.rts.GetDepartureTimes(ctx, *tp, *stn)
			if err != nil {
				if errors.Is(err, transit.ErrOperatorNotSupported) {
					continue
				}
				return nil, errors.Wrap(err, "transvc: get realtime departure times")
			}
			if len(times) == 0 {
				log.Warn("No realtime departure times found, skipping.")
				continue
			}

			nd := &nds[i]
			nd.Times = times
			nd.Realtime = true
		}
	}
	return nds, nil
}

var adjNumAlphaRegexp = regexp.MustCompile(`([0-9])([a-zA-Z])`)

func normalizeAdjacentNumAlpha(s string) string {
	return adjNumAlphaRegexp.ReplaceAllString(s, "$1 $2")
}
