package maps

import (
	"time"

	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/geo"
)

type (
	// A LocationService implements an api.LocationService for a Historian.
	LocationService struct {
		historian *Historian
		geocoder  geo.Geocoder
	}

	// An LSOption configures a LocationService.
	LSOption func(*LocationService)
)

var _ api.LocationService = (*LocationService)(nil)

// NewLocationService creates a new LocationService.
func NewLocationService(g geo.Geocoder, opts ...LSOption) (*LocationService,
	error) {
	svc := &LocationService{geocoder: g}
	for _, opt := range opts {
		opt(svc)
	}
	if svc.historian == nil {
		var err error
		if svc.historian, err = NewHistorian(); err != nil {
			return nil, errors.Errorf("maps: creating Historian: %w", err)
		}
	}
	return svc, nil
}

// WithLSHistorian configures a LocationService to fetch location history data
// with h.
func WithLSHistorian(h *Historian) LSOption {
	return func(svc *LocationService) { svc.historian = h }
}

// LastSeen gets the authenticated user's last seen coordinates.
func (ls *LocationService) LastSeen() (*geo.Coordinate, error) {
	// Get last seen coordinates.
	placemarks, err := ls.historian.LocationHistory(time.Now())
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

// CurrentCity calculates the authenticated user's current city.
func (ls *LocationService) CurrentCity() (city string, err error) {
	coord, err := ls.LastSeen()
	if err != nil {
		return "", errors.Errorf("maps: determining last seen position: %w", err)
	}
	if coord == nil {
		return "", errors.New("maps: no position data available")
	}
	return geo.CityAt(ls.geocoder, *coord)
}
