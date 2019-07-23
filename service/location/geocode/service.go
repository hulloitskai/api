package geocode

import (
	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/service/location"
)

type (
	// A LocationService implements an location.Service using a
	// location.HistoryService a Geocoder.
	LocationService struct {
		geocoder    Geocoder
		regionLevel Level

		history location.HistoryService
	}

	// A LocationServiceConfig configures a LocationService.
	LocationServiceConfig struct {
		RegionGeocodeLevel Level
	}
)

var _ location.Service = (*LocationService)(nil)

// NewLocationService creates a new LocationService.
func NewLocationService(
	history location.HistoryService,
	g Geocoder,
	opts ...func(*LocationServiceConfig),
) LocationService {
	cfg := LocationServiceConfig{RegionGeocodeLevel: CityLevel}
	for _, opt := range opts {
		opt(&cfg)
	}
	return LocationService{
		geocoder:    g,
		regionLevel: cfg.RegionGeocodeLevel,
		history:     history,
	}
}

// CurrentCity returns my current city.
func (svc LocationService) CurrentCity() (city string, err error) {
	// Get last position.
	coord, err := svc.LastPosition()
	if err != nil {
		return "", errors.Wrap(err, "geocode: getting last position")
	}
	if coord == nil {
		return "", errors.New("geocode: no position data available")
	}

	// Reverse-geocode coordinates.
	results, err := svc.geocoder.ReverseGeocode(
		*coord,
		func(cfg *ReverseGeocodeConfig) { cfg.Level = CityLevel },
	)
	if err != nil {
		return "", errors.Wrap(err, "geocode: reverse-geocoding last position")
	}
	if len(results) == 0 {
		return "", errors.New("geocode: no locations found at given position")
	}

	// Return city name.
	return results[0].Place.Address.Label, nil
}

// LastPosition returns my last seen position.
func (svc LocationService) LastPosition() (*location.Coordinates, error) {
	// Get recent history.
	segments, err := svc.RecentHistory()
	if err != nil {
		return nil, errors.Wrap(err, "geocode: getting recent location history")
	}
	if len(segments) == 0 {
		return nil, nil
	}

	segment := segments[len(segments)-1]
	if segment == nil {
		return nil, nil
	}
	var (
		coords = segment.Coordinates
		copy   = coords[len(coords)-1]
	)
	return &copy, nil
}

// CurrentRegion returns my current region.
func (svc LocationService) CurrentRegion() (*location.Place, error) {
	coord, err := svc.LastPosition()
	if err != nil {
		return nil, errors.Wrap(err, "geocode: determining last seen position")
	}
	if coord == nil {
		return nil, errors.New("geocode: no position data available")
	}

	results, err := svc.geocoder.ReverseGeocode(
		*coord,
		func(cfg *ReverseGeocodeConfig) {
			cfg.Level = svc.regionLevel
			cfg.IncludeShape = true
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "geocode: reverse-geocoding position")
	}
	if len(results) == 0 {
		return nil, errors.New("geocode: no locations found at given position")
	}
	return &results[0].Place, nil
}

// RecentHistory returns my recent location history.
func (svc LocationService) RecentHistory() ([]*location.HistorySegment, error) {
	return svc.history.RecentHistory()
}
