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

	"go.stevenxie.me/gopkg/logutil"
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
		logutil.MethodKey: name.OfMethod(service.FindDepartures),
		"route_query":     routeQuery,
		"coordinates":     coords,
	}).WithContext(ctx)

	// Validate inputs.
	if routeQuery == "" {
		return nil, errors.New("transvc: route is empty")
	}

	cfg := transit.FindDeparturesConfig{
		PreferRealtime: true,
		TimesLimit:     3,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "transvc: validating config")
	}
	{
		fields := logrus.Fields{
			"prefer_realtime":  cfg.PreferRealtime,
			"fuzzy_match":      cfg.FuzzyMatch,
			"group_by_station": cfg.GroupByStation,
			"limit":            cfg.Limit,
		}
		if c := cfg.OperatorCode; c != "" {
			fields["operator_code"] = c
		}
		log = log.WithFields(fields)
	}

	log.Trace("Getting nearby departures...")
	nds, err := svc.loc.NearbyDepartures(
		ctx,
		coords,
		func(ndCfg *transit.NearbyDeparturesConfig) {
			ndCfg.MaxPerTransport = cfg.TimesLimit
			if r := cfg.Radius; r > 0 {
				ndCfg.Radius = r
			}
			if m := cfg.MaxStations; m > 0 {
				ndCfg.MaxStations = m
			}
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to get nearby departures.")
		return nil, errors.Wrap(err, "transvc: get nearby departures")
	}
	log.WithField("departures", nds).Trace("Got nearby departures.")

	// Filter based on operator, if applicable.
	if code := cfg.OperatorCode; code != "" {
		var filtered []transit.NearbyDeparture
		for i := range nds {
			if nds[i].Transport.Operator.Code == code {
				filtered = append(filtered, nds[i])
			}
		}
		nds = filtered
		log.WithFields(logrus.Fields{
			"code":     code,
			"filtered": filtered,
		}).Trace("Filtered departures by operator code.")
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

		log := log.WithField("route", route)
		log.
			WithField("targets", routesWithContext).
			Trace("Performing fuzzy search against targets.")
		matches := fuzzy.RankFindFold(route, routesWithContext)
		if len(matches) == 0 {
			log.Trace("No matching routes, aborting.")
			return nil, errors.WithDetailf(
				errors.New("transvc: no matching route"),
				"No nearby routes matching '%s'.", routeQuery,
			)
		}
		sort.Sort(matches)
		log.
			WithField("matches", matches).
			Trace("Got fuzzy search matches, selecting first match.")
		route = routes[matches[0].OriginalIndex]
	} else {
		route = routeQuery
	}
	log = log.WithField("route", route)
	log.Trace("Derived route.")

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
			return nil, errors.WithDetailf(
				errors.New("transvc: no departures matching route"),
				"No departures found for '%s'.", routeQuery,
			)
		}
		nds = filtered
		log.
			WithField("filtered", filtered).
			Trace("Filtered results by route.")
	}

	// Group by station, if enabled.
	if cfg.GroupByStation {
		var (
			sids    = make(map[string]zero.Struct)
			grouped = make([]transit.NearbyDeparture, 0, len(nds))
		)
		for i := range nds {
			var (
				stn = nds[i].Station
				sid = stn.ID
			)
			if _, ok := sids[sid]; !ok {
				grouped = append(grouped, nds[i])
				sids[sid] = zero.Empty()

				// Find other matching stations by name.
				for j := i + 1; j < len(nds); j++ {
					otherStn := nds[j].Station
					if otherStn.Name == stn.Name {
						grouped = append(grouped, nds[j])
						sids[otherStn.ID] = zero.Empty()
					}
				}
			}
		}
		nds = grouped
		log.
			WithField("grouped", grouped).
			Trace("Grouped results by station.")
	}

	// Filter to a single set that is unique by direction.
	if cfg.SingleSet {
		var (
			dirs     = make(map[string]zero.Struct)
			filtered = make([]transit.NearbyDeparture, 0, len(nds))
		)
		for i := range nds {
			dir := nds[i].Transport.Direction
			if _, ok := dirs[dir]; ok {
				continue
			}
			dirs[dir] = zero.Empty()
			filtered = append(filtered, nds[i])
		}
		nds = filtered
		log.
			WithField("filtered", filtered).
			Trace("Filtered results to a single set by direction.")
	}

	// Apply limit.
	if l := cfg.Limit; (l > 0) && (len(nds) > l) {
		nds = nds[:l]
		log.WithFields(logrus.Fields{
			"limit":      l,
			"departures": nds,
		}).Trace("Applied departures limit.")
	}

	// Update with realtime departures times, if available.
	if cfg.PreferRealtime {
		log.Trace("Updating results with realtime data...")
		var modified bool
		for i := range nds {
			if nds[i].Realtime {
				continue // departure already is realtime
			}
			modified = true

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
				log.Info("No realtime departure times found, skipping.")
				continue
			}
			log.WithField("times", times).Trace("Got realtime departures.")

			nd := &nds[i]
			nd.Times = times
			nd.Realtime = true
		}
		if modified {
			log.
				WithField("departures", nds).
				Trace("Results were updated with realtime data.")
		} else {
			log.Trace("All pre-existing results are realtime; no modifications made.")
		}
	}
	return nds, nil
}

var adjNumAlphaRegexp = regexp.MustCompile(`([0-9])([a-zA-Z])`)

func normalizeAdjacentNumAlpha(s string) string {
	return adjNumAlphaRegexp.ReplaceAllString(s, "$1 $2")
}
