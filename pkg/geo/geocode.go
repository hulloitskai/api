package geo

import (
	errors "golang.org/x/xerrors"
)

type (
	// A Geocoder can look up geographical features that correspond to a set of
	// coordinates.
	Geocoder interface {
		ReverseGeocode(coord Coordinate, opts ...GeocodeOption) ([]*Feature, error)
	}

	// GeocodeOptions are a set of configurable options for a geocoding request.
	GeocodeOptions struct {
		Types []FeatureType
	}

	// A GeocodeOption configures a geocoding request.
	GeocodeOption func(*GeocodeOptions)
)

// WithTypes configures a geocoding request to limit the types of features
// to search for.
func WithTypes(types ...FeatureType) GeocodeOption {
	return func(opts *GeocodeOptions) { opts.Types = types }
}

// A FeatureType represents the type of a feature.
type FeatureType string

// A set of possible FeatureTypes.
const (
	CountryType      FeatureType = "country"
	RegionType       FeatureType = "region"
	PostcodeType     FeatureType = "postcode"
	DistrictType     FeatureType = "district"
	PlaceType        FeatureType = "place"
	LocalityType     FeatureType = "locality"
	NeighborhoodType FeatureType = "neighborhood"
	AddressType      FeatureType = "address"
	POIType          FeatureType = "poi"
)

type (
	// A Feature is a geographical feature.
	Feature struct {
		ID         string            `json:"id"`
		Type       []FeatureType     `json:"place_type"`
		Relevance  float32           `json:"relevance"`
		Text       string            `json:"text"`
		Place      string            `json:"place_name"`
		Properties FeatureProperties `json:"properties"`
		Context    []FeatureContext  `json:"context"`
	}

	// FeatureProperties describe the properties of a Feature.
	FeatureProperties struct {
		Landmark  bool   `json:"landmark"`
		Address   string `json:"address"`
		Category  string `json:"category"`
		Wikidata  string `json:"wikidata"`
		ShortCode string `json:"short_code"`
	}

	// FeatureContext describes the context of a Feature in terms of other related
	// features.
	FeatureContext struct {
		ID        string `json:"id"`
		ShortCode string `json:"short_code"`
		Wikidata  string `json:"wikidata"`
		Text      string `json:"text"`
	}
)

// CityAt uses a Geocoder to determine the city located at coord.
func CityAt(geo Geocoder, coord Coordinate) (city string, err error) {
	features, err := geo.ReverseGeocode(coord, WithTypes(PlaceType))
	if err != nil {
		return "", errors.Errorf("geo: reverse-geocoding last seen position: %w",
			err)
	}

	if len(features) == 0 {
		return "", errors.New("geo: no features found at last seen position")
	}
	return features[0].Place, nil
}
