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
			routeQuery string,
			coords location.Coordinates,
			opts ...FindDeparturesOption,
		) ([]NearbyDeparture, error)

		// NearbyTransports reports active Transports near a particular position.
		NearbyTransports(
			ctx context.Context,
			coords location.Coordinates,
			opts ...NearbyTransportsOption,
		) ([]Transport, error)
	}

	// A FindDeparturesConfig configures a Service.FindDepartures request.
	FindDeparturesConfig struct {
		GroupByStation bool   // group results by station
		FuzzyMatch     bool   // use fuzzy match algorithm for route
		OperatorCode   string // filter by operator code
		Limit          int    // limit number of results
		MaxTimes       int    // maximum number of departure times to include
		Radius         *int   // the search radius, in meters
		MaxStations    *int   // max number of stations to search
	}

	// A FindDeparturesOption modifies a FindDeparturesConfig.
	FindDeparturesOption func(*FindDeparturesConfig)

	// A NearbyTransportsConfig configures a Service.NearbyTransports request.
	NearbyTransportsConfig struct {
		Radius      *int
		Limit       *int
		MaxStations *int
	}

	// A NearbyTransportsOption modifies a NearbyTransportsConfig.
	NearbyTransportsOption func(*NearbyTransportsConfig)
)

// FindWithGroupByStation enables the grouping of results by station for a
// Service.FindDepartures request.
func FindWithGroupByStation(enable bool) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.FuzzyMatch = enable }
}

// FindWithFuzzyMatch enables fuzzy-matching on the route for a
// Service.FindDepartures request.
func FindWithFuzzyMatch(enable bool) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.FuzzyMatch = enable }
}

// FindWithRadius configures a Service.FindDepartures request to limit search
// to departures within r meters of the provided position.
func FindWithRadius(r int) FindDeparturesOption {
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
