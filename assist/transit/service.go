package transit // import "go.stevenxie.me/api/v2/assist/transit"

import (
	"context"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation"
	"go.stevenxie.me/api/v2/location"
)

// FindWithGroupByStation enables the grouping of results by station for a
// Service.NearbyDepartures request.
func FindWithGroupByStation(enable bool) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) { opt.FuzzyMatch = enable }
}

// FindWithFuzzyMatch enables fuzzy-matching on the route for a
// Service.NearbyDepartures request.
func FindWithFuzzyMatch(enable bool) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) { opt.FuzzyMatch = enable }
}

// FindWithRadius configures a Service.NearbyDepartures request to limit search
// to departures within r meters of the provided position.
func FindWithRadius(r int) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) {
		if r > 0 {
			opt.Radius = r
		}
	}
}

// FindWithOperator restricts the search to the Operator with the specified
// code.
func FindWithOperator(opCode string) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) { opt.OperatorCode = opCode }
}

// FindWithLimit limits the number of results from a Service.NearbyDepartures
// request.
func FindWithLimit(l int) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) {
		if l > 0 {
			opt.Limit = l
		}
	}
}

// FindSingleSet instructs a Service.NearbyDepartures call to only include
// a single set of results that are unique by direction.
func FindSingleSet(enable bool) NearbyDeparturesOption {
	return func(opt *NearbyDeparturesOptions) { opt.SingleSet = enable }
}

type (
	// A Service can assist me with my transit needs.
	Service interface {
		// NearbyDepartures finds departures for a particular transit route near
		// pos.
		NearbyDepartures(
			ctx context.Context,
			pos location.Coordinates,
			routeQuery string,
			opts ...NearbyDeparturesOption,
		) ([]NearbyDeparture, error)

		// NearbyTransports finds active Transports near a particular position.
		NearbyTransports(
			ctx context.Context,
			pos location.Coordinates,
			opts ...NearbyTransportsOption,
		) ([]Transport, error)
	}

	// A NearbyDeparturesOptions are option parameters for a
	// Service.NearbyDepartures request.
	NearbyDeparturesOptions struct {
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

	// A NearbyDeparturesOption modifies a NearbyDeparturesOptions.
	NearbyDeparturesOption func(*NearbyDeparturesOptions)

	// NearbyTransportsOptions are option parameters for a
	// Service.NearbyTransports request.
	NearbyTransportsOptions struct {
		Radius      int
		Limit       int
		MaxStations int
	}

	// A NearbyTransportsOption modifies a NearbyTransportsOption.
	NearbyTransportsOption func(*NearbyTransportsOptions)
)

// Validate returns an error if the FindDeparrturesOption is not valid.
func (opt *NearbyDeparturesOptions) Validate() error {
	minZeroFields := []*int{
		&opt.Limit, &opt.TimesLimit,
		&opt.Radius, &opt.MaxStations,
	}
	rules := make([]*validation.FieldRules, len(minZeroFields))
	for i, f := range minZeroFields {
		rules[i] = validation.Field(f, validation.Min(0))
	}
	if err := validation.ValidateStruct(opt, rules...); err != nil {
		return err
	}

	if l := opt.Limit; opt.SingleSet && l > 0 {
		return errors.Newf("Limit (%d) cannot be set when SingleSet is true", l)
	}
	return nil
}
