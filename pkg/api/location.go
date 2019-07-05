package api

import (
	"github.com/stevenxie/api/pkg/geo"
)

// A LocationService provides information about my recent locations.
type LocationService interface {
	LastSeen() (*geo.Coordinate, error)
	CurrentCity() (city string, err error)
}
