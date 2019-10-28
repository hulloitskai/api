package geocode

import (
	"context"

	"go.stevenxie.me/api/v2/location"
)

// A Geocoder can look up geographical features that correspond to a set of
// coordinates.
type Geocoder interface {
	ReverseGeocode(
		ctx context.Context,
		coord location.Coordinates,
		opts ...ReverseGeocodeOption,
	) ([]ReverseGeocodeResult, error)
}
