package geocode

import (
	"go.stevenxie.me/api/location"
)

type (
	// ReverseGeocodeConfig are a set of configurable options for a
	// reverse-geocoding request.
	ReverseGeocodeConfig struct {
		Level           Level
		Radius          uint
		IncludeShape    bool
		IncludeTimeZone bool
	}

	// A ReverseGeocodeOption modifies a ReverseGeocodeConfig.
	ReverseGeocodeOption func(*ReverseGeocodeConfig)

	// A ReverseGeocodeResult is the result of a reverse-geocoding search.
	ReverseGeocodeResult struct {
		Place     location.Place
		Relevance float32
		Distance  float32
	}
)

// WithReverseGeocodeLevel sets the geocoding level (proximity) of a
// reverse-geocoding request.
func WithReverseGeocodeLevel(l Level) ReverseGeocodeOption {
	return func(cfg *ReverseGeocodeConfig) { cfg.Level = l }
}

// WithReverseGeocodeRadius sets the search radius of a reverse-geocoding
// request.
func WithReverseGeocodeRadius(radius uint) ReverseGeocodeOption {
	return func(cfg *ReverseGeocodeConfig) { cfg.Radius = radius }
}

// IncludeReverseGeocodeShape sets a reverse-geocoding request to include
// geographical shape information in the response.
func IncludeReverseGeocodeShape() ReverseGeocodeOption {
	return func(cfg *ReverseGeocodeConfig) { cfg.IncludeShape = true }
}

// IncludeReverseGeocodeTimeZone sets a reverse-geocoding request to include
// time zone information in the response.
func IncludeReverseGeocodeTimeZone() ReverseGeocodeOption {
	return func(cfg *ReverseGeocodeConfig) { cfg.IncludeTimeZone = true }
}
