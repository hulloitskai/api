package geo

import (
	"fmt"

	errors "golang.org/x/xerrors"
)

type (
	// A Geocoder can look up geographical features that correspond to a set of
	// coordinates.
	Geocoder interface {
		ReverseGeocode(coord Coordinate, opts ...RGOption) ([]*RGResult, error)
	}

	// A RGResult is the result of a reverse-geocoding search.
	RGResult struct {
		Location  `json:"location"`
		Relevance float32 `json:"relevance"`
		Distance  float32 `json:"distance"`
	}

	// RGOptions are a set of configurable options for a reverse-geocoding
	// request.
	RGOptions struct {
		Level        GeocodeLevel
		Radius       uint
		IncludeShape bool
	}

	// A RGOption configures a reverse-geocoding request.
	RGOption func(*RGOptions)
)

// WithRGLevel configures a reverse-geocoding request to limit the search scope
// specified geocoding match level.
func WithRGLevel(l GeocodeLevel) RGOption {
	return func(opts *RGOptions) { opts.Level = l }
}

// WithRGRadius configures a reverse-geocoding request to limit the search scope
// to the specified radius.
func WithRGRadius(radius uint) RGOption {
	return func(opts *RGOptions) { opts.Radius = radius }
}

// WithRGShape confiures a reverse-geocoding request to include an area shape.
func WithRGShape() RGOption {
	return func(opts *RGOptions) { opts.IncludeShape = true }
}

// A GeocodeLevel represents the type of a feature.
type GeocodeLevel uint8

// A set of possible GeocodeLevels.
const (
	CountryLevel GeocodeLevel = iota + 1
	StateLevel
	CountyLevel
	CityLevel
	DistrictLevel
	PostcodeLevel
)

var geocodeLevelNames = map[GeocodeLevel]string{
	CountryLevel:  "country",
	StateLevel:    "state",
	CountyLevel:   "county",
	CityLevel:     "city",
	DistrictLevel: "district",
	PostcodeLevel: "postcode",
}

func (level GeocodeLevel) String() string {
	if name, ok := geocodeLevelNames[level]; ok {
		return name
	}
	return fmt.Sprintf("GeocodeLevel(%d)", uint8(level))
}

type (
	// A Location is a geographical location.
	Location struct {
		ID       string       `json:"id"`
		Level    string       `json:"level"`
		Type     string       `json:"type"`
		Position Coordinate   `json:"position"`
		Address  *Address     `json:"address"`
		Shape    []Coordinate `json:"shape,omitempty"`
	}

	// An Address describes the position of a location.
	Address struct {
		Label      string `json:"label"`
		Country    string `json:"country"`
		State      string `json:"state"`
		County     string `json:"county"`
		City       string `json:"city"`
		District   string `json:"district,omitempty"`
		PostalCode string `json:"postalCode"`
		Street     string `json:"street,omitempty"`
		Number     string `json:"number,omitempty"`
	}
)

// CityAt uses a Geocoder to determine the city located at coord.
func CityAt(geo Geocoder, coord Coordinate) (city string, err error) {
	results, err := geo.ReverseGeocode(coord, WithRGLevel(CityLevel))
	if err != nil {
		return "", errors.Errorf("geo: reverse-geocoding position: %w", err)
	}
	if len(results) == 0 {
		return "", errors.New("geo: no locations found at given position")
	}
	addr := results[0].Address
	return fmt.Sprintf("%s, %s, %s", addr.County, addr.State, addr.Country), nil
}

// RegionAt uses a Geocoder to look up the region that coord is situated in.
//
// It will also request for the shape of the region area.
func RegionAt(geo Geocoder, coord Coordinate) (*Location, error) {
	results, err := geo.ReverseGeocode(
		coord,
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
