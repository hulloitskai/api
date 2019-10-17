package transgql

import (
	"context"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/location/locgql"
)

// NewQuery creates a new Query.
func NewQuery(svc transit.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for transit-related data.
type Query struct {
	svc transit.Service
}

// FindDepartures forwards a definition.
func (q Query) FindDepartures(
	ctx context.Context,
	route string,
	near locgql.CoordinatesInput,
	limit *int,
) ([]transit.NearbyDeparture, error) {
	lim := 2
	if limit != nil {
		lim = *limit
	}
	return q.svc.FindDepartures(
		ctx,
		route, locgql.CoordinatesFromInput(near),
		transit.FindWithFuzzyMatch(true),
		transit.FindWithLimit(lim),
	)
}
