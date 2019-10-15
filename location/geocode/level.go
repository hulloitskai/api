package geocode

import (
	stderrs "errors"
	"fmt"
)

// A Level represents the type of a feature.
type Level uint8

// A set of possible GeocodeLevels.
const (
	CountryLevel Level = iota + 1
	StateLevel
	CountyLevel
	CityLevel
	DistrictLevel
	PostcodeLevel
)

var levelNames = map[Level]string{
	CountryLevel:  "Country",
	StateLevel:    "State",
	CountyLevel:   "County",
	CityLevel:     "City",
	DistrictLevel: "District",
	PostcodeLevel: "Postcode",
}

func (level Level) String() string {
	if name, ok := levelNames[level]; ok {
		return name
	}
	return fmt.Sprintf("Level(%d)", uint8(level))
}

// ParseLevel parses a string representing a geocode level into a
// Level.
func ParseLevel(s string) (Level, error) {
	for key, val := range levelNames {
		if val == s {
			return key, nil
		}
	}
	return 0, ErrInvalidLevel
}

// ErrInvalidLevel is an error returned by ParseLevel when no
// matching geocode level name is found.
var ErrInvalidLevel = stderrs.New("geocode: invalid geocode level")
