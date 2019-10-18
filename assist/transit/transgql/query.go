package transgql

import (
	"context"

	"github.com/cockroachdb/errors"
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
	radius *int,
	limit *int,
) ([]transit.NearbyDeparture, error) {
	// Marshal input parameters.
	var (
		lim = 2
		rad *uint
	)
	if limit != nil {
		lim = *limit
	}
	if radius != nil {
		if *radius < 0 {
			return nil, errors.New("transgql: radius may not be negative")
		}
		r := uint(*radius)
		rad = &r
	}
	return q.svc.FindDepartures(
		ctx,
		route, locgql.CoordinatesFromInput(near),
		func(cfg *transit.FindDeparturesConfig) {
			cfg.GroupByStation = true
			cfg.FuzzyMatch = true
			cfg.Limit = lim
			if rad != nil {
				cfg.Radius = rad
			}
		},
	)
}
