package heregeo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/api/pkg/here"
)

// NewGeocoder creates a new geocode.Geocoder.
func NewGeocoder(c here.Client) geocode.Geocoder {
	return geocoder{
		client: c,
	}
}

type geocoder struct {
	client here.Client
}

var _ geocode.Geocoder = (*geocoder)(nil)

func (g geocoder) ReverseGeocode(
	ctx context.Context,
	coord location.Coordinates,
	opts ...geocode.ReverseGeocodeOption,
) ([]geocode.ReverseGeocodeResult, error) {
	// Build and validate config.
	var cfg geocode.ReverseGeocodeConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := validateReverseGeocodeConfig(&cfg); err != nil {
		return nil, errors.Wrap(err, "heregeo: invalid config")
	}

	// Create and perform request.
	url := buildReverseGeocodeURL(coord, &cfg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "heregeo: creating request")
	}
	res, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode response.
	var data struct {
		Response struct {
			View []struct {
				Result []struct {
					Relevance  float32
					Distance   float32
					MatchLevel string
					Location   struct {
						ID       string `json:"LocationId"`
						Type     string `json:"LocationType"`
						Position struct {
							Latitude  float64
							Longitude float64
						} `json:"DisplayPosition"`
						Address struct {
							Label       string
							Country     string
							State       string
							County      string
							City        string
							District    string
							PostalCode  string
							Street      string
							HouseNumber string
						}
						Shape *struct {
							Value string
						}
						AdminInfo *struct {
							TimeZone struct {
								ID string `json:"id"`
							}
						}
					}
				}
			}
		}
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "heregeo: decoding response body")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "heregeo: closing response body")
	}

	// Parse response.
	if len(data.Response.View) == 0 {
		return nil, errors.New("heregeo: no result views")
	}

	var (
		matches = data.Response.View[0].Result
		results = make([]geocode.ReverseGeocodeResult, len(matches))
	)
	for i, match := range matches {
		var (
			loc  = &match.Location
			pos  = &loc.Position
			addr = &loc.Address
		)

		// Decode shape in response.
		var shape []location.Coordinates
		if res := loc.Shape; res != nil {
			if shape, err = decodeShapeResponse(res.Value); err != nil {
				return nil, errors.Wrap(err, "heregeo: decoding shape")
			}
		}

		// Load timezone.
		var timeZone *time.Location
		if info := loc.AdminInfo; info != nil {
			if timeZone, err = time.LoadLocation(info.TimeZone.ID); err != nil {
				return nil, errors.Wrap(err, "heregeo: parsing timezone")
			}
		}

		// Save result.
		results[i] = geocode.ReverseGeocodeResult{
			Place: location.Place{
				ID:    loc.ID,
				Level: match.MatchLevel,
				Type:  loc.Type,
				Position: location.Coordinates{
					X: pos.Longitude,
					Y: pos.Latitude,
				},
				Address: location.Address{
					Label:    addr.Label,
					Country:  addr.Country,
					State:    addr.State,
					County:   addr.County,
					City:     addr.City,
					District: addr.District,
					Postcode: addr.PostalCode,
					Street:   addr.Street,
					Number:   addr.HouseNumber,
				},
				TimeZone: timeZone,
				Shape:    shape,
			},
		}
	}
	return results, nil
}

func validateReverseGeocodeConfig(cfg *geocode.ReverseGeocodeConfig) error {
	if cfg.IncludeShape && (cfg.Level == 0) {
		return errors.New("cannot include area shape without level selection")
	}
	return nil
}

const _reverseGeocodeURL = "https://reverse.geocoder.api.here.com/6.2/" +
	"reversegeocode.json"

func buildReverseGeocodeURL(
	coord location.Coordinates,
	cfg *geocode.ReverseGeocodeConfig,
) string {
	// Build request URL.
	url, err := url.Parse(_reverseGeocodeURL)
	if err != nil {
		panic(err)
	}

	params := url.Query()
	params.Set("gen", "9")
	params.Set("mode", "retrieveAll")

	// Set location attributes.
	{
		attrs := []string{"address"}
		if cfg.IncludeTimeZone {
			attrs = append(attrs, "timeZone")
		}
		params.Set("locationattributes", strings.Join(attrs, ","))
	}
	if cfg.IncludeShape {
		params.Set("additionalData", "IncludeShapeLevel,default")
	}

	// Set geocoding proximity.
	{
		var radius uint = 50
		if cfg.Radius > 0 {
			radius = cfg.Radius
		}
		params.Set("prox", fmt.Sprintf("%f,%f,%d", coord.Y, coord.X, radius))
	}

	// Set geocoding level.
	if cfg.Level > 0 {
		level := strings.ToLower(cfg.Level.String())
		if cfg.Level == geocode.PostcodeLevel {
			level = "postalCode"
		}
		params.Set("level", level)
	}

	// Encode params and build URL.
	url.RawQuery = params.Encode()
	return url.String()
}

func decodeShapeResponse(value string) (shape []location.Coordinates,
	err error) {
	if !strings.HasPrefix(value, "POLYGON") {
		return nil, nil
	}

	// Split value into coordinate pairs.
	value = value[10 : len(value)-2]
	pairs := strings.Split(value, ", ")

	// Parse pairs into location.Coordinates.
	for _, pair := range pairs {
		var (
			coords = strings.Fields(pair)
			floats = make([]float64, len(coords))
		)
		for i, coord := range coords {
			if floats[i], err = strconv.ParseFloat(coord, 64); err != nil {
				return nil, errors.Wrapf(
					err,
					"parsing partial-coordinate '%s'", coord,
				)
			}
		}
		shape = append(shape, location.Coordinates{
			X: floats[0],
			Y: floats[1],
		})
	}
	return shape, nil
}
