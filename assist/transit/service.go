package transit

import (
	"context"

	"go.stevenxie.me/api/location"
)

type (
	// A Service can assist me with my transit needs.
	Service interface {
		// FindDepartures finds departures for a particular transit route near
		// pos.
		FindDepartures(
			ctx context.Context,
			route string,
			pos location.Coordinates,
			opts ...FindDeparturesOption,
		) ([]NearbyDeparture, error)
	}

	// A FindDeparturesConfig configures a Service.FindDepartures request.
	FindDeparturesConfig struct {
		OperatorCode string // filter by operator code
		FuzzyMatch   bool   // use fuzzy match algorithm for route
		Radius       *uint  // the search radius, in meters
		MaxStations  *uint  // max number of stations to search
		Limit        int    // limit number of results
	}

	// A FindDeparturesOption modifies a FindDeparturesConfig.
	FindDeparturesOption func(*FindDeparturesConfig)
)

// FindWithFuzzyMatch enables fuzzy-matching on the route for a
// Service.FindDepartures request.
func FindWithFuzzyMatch() FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.FuzzyMatch = true }
}

// FindWithRadius configures a Service.FindDepartures request to limit search
// to departures within r meters of the provided position.
func FindWithRadius(r uint) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.Radius = &r }
}

// RestrictFindOperator restricts the search to the Operator with the specified
// code.
func RestrictFindOperator(opCode string) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.OperatorCode = opCode }
}

// FindWithLimit limits the number of results from a Service.FindDepartures
// request.
func FindWithLimit(l int) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.Limit = l }
}
