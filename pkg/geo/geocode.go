package geo

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type (
	// A Geocoder can look up geographical features that correspond to a set of
	// coordinates.
	Geocoder interface {
		ReverseGeocode(
			coord Coordinate,
			opts ...func(*ReverseGeocodeConfig),
		) ([]*ReverseGeocodeResult, error)
	}

	// A ReverseGeocodeResult is the result of a reverse-geocoding search.
	ReverseGeocodeResult struct {
		Location  `json:"location"`
		Relevance float32 `json:"relevance"`
		Distance  float32 `json:"distance"`
	}

	// ReverseGeocodeConfig are a set of configurable options for a
	// reverse-geocoding request.
	ReverseGeocodeConfig struct {
		Level        GeocodeLevel
		Radius       uint
		IncludeShape bool
	}
)

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
	CountryLevel:  "Country",
	StateLevel:    "State",
	CountyLevel:   "County",
	CityLevel:     "City",
	DistrictLevel: "District",
	PostcodeLevel: "Postcode",
}

func (level GeocodeLevel) String() string {
	if name, ok := geocodeLevelNames[level]; ok {
		return name
	}
	return fmt.Sprintf("GeocodeLevel(%d)", uint8(level))
}

// ParseGeocodeLevel parses a string representing a geocode level into a
// GeocodeLevel.
func ParseGeocodeLevel(s string) (GeocodeLevel, error) {
	for key, val := range geocodeLevelNames {
		if val == s {
			return key, nil
		}
	}
	return 0, ErrInvalidGeocodeLevel
}

// ErrInvalidGeocodeLevel is an error returned by ParseGeocodeLevel when no
// matching geocode level name is found.
var ErrInvalidGeocodeLevel = errors.New("geo: invalid geocode level")

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
