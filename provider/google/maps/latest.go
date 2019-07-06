package maps

import (
	"time"

	"github.com/stevenxie/api/pkg/geo"
	errors "golang.org/x/xerrors"
)

// LatestCoordinates returns the authenticated user's latest coordinates.
func (h *Historian) LatestCoordinates() (*geo.Coordinate, error) {
	placemarks, err := h.LocationHistory(time.Now())
	if err != nil {
		return nil, errors.Errorf("maps: fetching location history: %w", err)
	}

	// Reverse-iterate through placemarks.
	for i := len(placemarks) - 1; i >= 0; i-- {
		if len(placemarks[i].Coordinates) == 0 {
			continue
		}
		coords := placemarks[i].Coordinates
		return &coords[len(coords)-1], nil
	}

	return nil, nil
}
