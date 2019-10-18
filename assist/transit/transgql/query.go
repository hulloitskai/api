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
	coords locgql.CoordinatesInput,
	radius *int,
	limit *int,
) ([]transit.NearbyDeparture, error) {
	// Marshal input parameters.
	lim := 2
	if limit != nil {
		lim = *limit
	}
	return q.svc.FindDepartures(
		ctx,
		route, locgql.CoordinatesFromInput(coords),
		func(cfg *transit.FindDeparturesConfig) {
			cfg.GroupByStation = true
			cfg.FuzzyMatch = true
			cfg.Limit = lim
			if radius != nil {
				cfg.Radius = radius
			}
		},
	)
}

// NearbyTransports forwards a definition.
func (q Query) NearbyTransports(
	ctx context.Context,
	coords locgql.CoordinatesInput,
	radius *int,
	limit *int,
) ([]transit.Transport, error) {
	return q.svc.NearbyTransports(
		ctx,
		locgql.CoordinatesFromInput(coords),
		func(cfg *transit.NearbyTransportsConfig) {
			if radius != nil {
				cfg.Radius = radius
			}
			if limit != nil {
				cfg.Limit = limit
			}
		},
	)
}
