package transit // import "go.stevenxie.me/api/assist/transit"

import (
	"context"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"go.stevenxie.me/api/location"
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
	return func(cfg *FindDeparturesConfig) {
		if r > 0 {
			cfg.Radius = r
		}
	}
}

// FindWithOperator restricts the search to the Operator with the specified
// code.
func FindWithOperator(opCode string) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.OperatorCode = opCode }
}

// FindWithLimit limits the number of results from a Service.FindDepartures
// request.
func FindWithLimit(l int) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) {
		if l > 0 {
			cfg.Limit = l
		}
	}
}

// FindSingleSet instructs a Service.FindDepartures call to only include
// a single set of results that are unique by direction.
func FindSingleSet(enable bool) FindDeparturesOption {
	return func(cfg *FindDeparturesConfig) { cfg.SingleSet = enable }
}

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
		OperatorCode string // filter by operator code

		Realtime   bool // make extra queries for realtime data
		FuzzyMatch bool // use fuzzy match algorithm for route

		GroupByStation bool // group results by station
		SingleSet      bool // get one result per direction (overrides Limit)

		Limit      int // limit number of results
		TimesLimit int // limit number of departure times to include

		Radius      int // the search radius, in meters
		MaxStations int // max number of stations to search
	}

	// A FindDeparturesOption modifies a FindDeparturesConfig.
	FindDeparturesOption func(*FindDeparturesConfig)

	// A NearbyTransportsConfig configures a Service.NearbyTransports request.
	NearbyTransportsConfig struct {
		Radius      int
		Limit       int
		MaxStations int
	}

	// A NearbyTransportsOption modifies a NearbyTransportsConfig.
	NearbyTransportsOption func(*NearbyTransportsConfig)
)

// Validate returns an error if the FindDeparrturesConfig is not valid.
func (cfg *FindDeparturesConfig) Validate() error {
	minZeroFields := []*int{
		&cfg.Limit, &cfg.TimesLimit,
		&cfg.Radius, &cfg.MaxStations,
	}
	rules := make([]*validation.FieldRules, len(minZeroFields))
	for i, f := range minZeroFields {
		rules[i] = validation.Field(f, validation.Min(0))
	}
	if err := validation.ValidateStruct(cfg, rules...); err != nil {
		return err
	}

	if l := cfg.Limit; cfg.SingleSet && l > 0 {
		return errors.Newf("Limit (%d) cannot be set when SingleSet is true", l)
	}
	return nil
}
