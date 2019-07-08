package geo

import (
	"fmt"

	errors "golang.org/x/xerrors"
)

// A Coordinate represents a point in 3D space.
type Coordinate struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type (
	// A LocationService implements an api.LocationService using a
	// SegmentSource and a Geocoder.
	LocationService struct {
		source   SegmentSource
		geocoder Geocoder
	}

	// A SegmentSource can fetch recent location history segments.
	SegmentSource interface {
		RecentSegments() ([]*Segment, error)
	}
)

// NewLocationService creates a new LocationService.
func NewLocationService(
	source SegmentSource,
	g Geocoder,
) LocationService {
	return LocationService{
		source:   source,
		geocoder: g,
	}
}

// RecentSegments returns the authenticated user's recent location history.
func (svc LocationService) RecentSegments() ([]*Segment, error) {
	return svc.source.RecentSegments()
}

// LastSegment returns the authenticated user's latest location history segment.
func (svc LocationService) LastSegment() (*Segment, error) {
	segments, err := svc.RecentSegments()
	if err != nil {
		return nil, err
	}
	if len(segments) == 0 {
		return nil, nil
	}
	return segments[len(segments)-1], nil
}

// LastPosition returns the authenticated user's last known position.
func (svc LocationService) LastPosition() (*Coordinate, error) {
	segment, err := svc.LastSegment()
	if err != nil {
		return nil, errors.Errorf("geo: fetching last location history segment: %w",
			err)
	}
	if segment == nil {
		return nil, nil
	}
	var (
		coords = segment.Coordinates
		copy   = coords[len(coords)-1]
	)
	return &copy, nil
}

// CurrentCity returns the authenticated user's current city.
func (svc LocationService) CurrentCity() (city string, err error) {
	coord, err := svc.LastPosition()
	if err != nil {
		return "", errors.Errorf("geo: determining last seen position: %w", err)
	}
	if coord == nil {
		return "", errors.New("geo: no position data available")
	}

	results, err := svc.geocoder.ReverseGeocode(*coord, WithRGLevel(CityLevel))
	if err != nil {
		return "", errors.Errorf("geo: reverse-geocoding position: %w", err)
	}
	if len(results) == 0 {
		return "", errors.New("geo: no locations found at given position")
	}
	addr := results[0].Address
	return fmt.Sprintf("%s, %s, %s", addr.County, addr.State, addr.Country), nil
}

// CurrentRegion returns the authenticated user's current region.
func (svc LocationService) CurrentRegion() (*Location, error) {
	coord, err := svc.LastPosition()
	if err != nil {
		return nil, errors.Errorf("geo: determining last seen position: %w", err)
	}
	if coord == nil {
		return nil, errors.New("geo: no position data available")
	}

	results, err := svc.geocoder.ReverseGeocode(
		*coord,
		WithRGLevel(CityLevel),
		WithRGShape(),
	)
	if err != nil {
		return nil, errors.Errorf("geo: reverse-geocoding position: %w", err)
	}
	if len(results) == 0 {
		return nil, errors.New("geo: no locations found at given position")
	}
	return &results[0].Location, nil
}
