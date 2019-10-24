package geoutil

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
)

// TimeLocation returns the time.Location that corresponds to a particular
// location.Coordinates.
func TimeLocation(
	ctx context.Context,
	geo geocode.Geocoder,
	coords location.Coordinates,
) (*time.Location, error) {
	res, err := geo.ReverseGeocode(
		ctx,
		coords,
		geocode.ReverseWithTimeZone(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "geoutil: reverse-geocoding timezone info")
	}
	if len(res) == 0 {
		return nil, errors.New("geoutil: no reverse-geocoding results")
	}
	return res[0].Place.TimeZone, nil
}
