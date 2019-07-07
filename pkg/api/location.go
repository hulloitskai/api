package api

import (
	"github.com/stevenxie/api/pkg/geo"
)

type (
	// A LocationService provides information about my recent locations.
	LocationService interface {
		CurrentCity() (city string, err error)
		CurrentRegion() (*geo.Location, error)
		LastPosition() (*geo.Coordinate, error)
		LastSegment() (*geo.Segment, error)
		RecentSegments() ([]*geo.Segment, error)
	}

	// A LocationStreamingService can stream my current location.
	LocationStreamingService interface {
		LocationService
		SegmentsStream() <-chan struct {
			Segment *geo.Segment
			Err     error
		}
	}

	// A LocationAccessService can validate location access codes.
	LocationAccessService interface {
		IsValidCode(code string) (valid bool, err error)
	}
)
