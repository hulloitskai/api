package transvc

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/errors/exthttp"

	"github.com/cockroachdb/errors"
	"github.com/lithammer/fuzzysearch/fuzzy"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
	"go.stevenxie.me/gopkg/zero"

	"go.stevenxie.me/api/v2/assist/assistutil"
	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/location"
)

// FindDepartures implements transit.Service.FindDepartures.
func (svc *service) FindDepartures(
	ctx context.Context,
	routeQuery string,
	coords location.Coordinates,
	opts ...transit.FindDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc((*service).FindDepartures),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*service).FindDepartures),
		"route_query":     routeQuery,
		"coordinates":     coords,
	}).WithContext(ctx)

	// Validate inputs.
	if routeQuery == "" {
		return nil, errors.New("transvc: route is empty")
	}

	opt := transit.FindDeparturesOptions{
		Realtime:   true,
		TimesLimit: 3,
	}
	for _, apply := range opts {
		apply(&opt)
	}
	if err := opt.Validate(); err != nil {
		return nil, errors.Wrap(err, "transvc: validating config")
	}
	{
		fields := logrus.Fields{
			"realtime":         opt.Realtime,
			"fuzzy_match":      opt.FuzzyMatch,
			"group_by_station": opt.GroupByStation,
			"limit":            opt.Limit,
			"times_limit":      opt.TimesLimit,
		}
		if c := opt.OperatorCode; c != "" {
			fields["operator_code"] = c
		}
		log = log.WithFields(fields)
	}

	log.Trace("Getting nearby departures...")
	nds, err := svc.loc.NearbyDepartures(
		ctx,
		coords,
		func(ndOpt *transit.NearbyDeparturesOptions) {
			ndOpt.MaxPerTransport = opt.TimesLimit
			if r := opt.Radius; r > 0 {
				ndOpt.Radius = r
			}
			if m := opt.MaxStations; m > 0 {
				ndOpt.MaxStations = m
			}
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to get nearby departures.")
		return nil, errors.Wrap(err, "transvc: get nearby departures")
	}
	log.WithField("departures", nds).Trace("Got nearby departures.")

	// Filter based on operator, if applicable.
	if code := opt.OperatorCode; code != "" {
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
	if opt.FuzzyMatch {
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
			err = errors.New("transvc: no matching route")
			err = errors.WithDetailf(
				err,
				"No nearby routes matching '%s'.", routeQuery,
			)
			return nil, exthttp.WrapWithHTTPCode(err, http.StatusNotFound)
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
			if code := opt.OperatorCode; code != "" {
				if nds[i].Transport.Operator.Code != code {
					continue
				}
			}
			if route == nds[i].Transport.Route {
				filtered = append(filtered, nds[i])
			}
		}
		if len(filtered) == 0 {
			err = errors.New("transvc: no departures matching route")
			return nil, errors.WithDetailf(
				err,
				"No departures found for '%s'.", routeQuery,
			)
		}
		nds = filtered
		log.
			WithField("filtered", filtered).
			Trace("Filtered results by route.")
	}

	// Group by station, if enabled.
	if opt.GroupByStation {
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
	if opt.SingleSet {
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
	if l := opt.Limit; (l > 0) && (len(nds) > l) {
		nds = nds[:l]
		log.WithFields(logrus.Fields{
			"limit":      l,
			"departures": nds,
		}).Trace("Applied departures limit.")
	}

	// Update with realtime departures times, if available.
	if opt.Realtime {
		log.
			WithField("realtime_sources", svc.rts).
			Trace("Updating results with realtime data...")
		var (
			now      = time.Now()
			modified bool
		)
		for i := range nds {
			nd := &nds[i]
			if nd.Realtime {
				continue // departure already is realtime
			}
			if len(nd.Times) > 0 {
				if nd.Times[0].Sub(now) < svc.maxRTDepGap {
					log.WithFields(logrus.Fields{
						"departure_times":   nd.Times,
						"max_departure_gap": svc.maxRTDepGap,
					}).Debug("Departure times too distanct; skipping realtime update.")
				}
			}
			var (
				tp = nd.Transport
				op = tp.Operator
			)
			log := log.WithField("operator", op)
			rts, ok := svc.rts[op.Code]
			if !ok {
				log.Debug("No registered transit.RealtimeSource for this operator.")
				continue // operator not supported
			}

			stn := nd.Station
			log = log.WithFields(logrus.Fields{
				"direction": tp.Direction,
				"station":   stn.Name,
			})

			log.Trace("Getting realtime departure...")
			times, err := rts.GetDepartureTimes(ctx, *tp, *stn)
			if err != nil {
				log := log.WithError(err)
				if errors.Is(err, transit.ErrOperatorNotSupported) {
					log.Warn("Incorrect transit.RealtimeSource for this operator.")
					continue
				}
				log.Error("Failed to get realtime departure times.")
				continue
			}
			if len(times) == 0 {
				log.Info("No realtime departure times found, skipping.")
				continue
			}
			log.WithField("times", times).Trace("Got realtime departures.")

			// Modify transit.NearbyDeparture.
			nd.Times = times
			nd.Realtime = true
			modified = true
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
