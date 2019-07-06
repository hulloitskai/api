package here

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/stevenxie/api/pkg/geo"
	errors "golang.org/x/xerrors"
)

const reverseGeocodeURL = "https://reverse.geocoder.api.here.com/6.2/" +
	"reversegeocode.json"

// ReverseGeocode reverse-geocodes a coordinate.
//
// Note that the geo.WithRGShape option can only be used if the geo.WithRGLevel
// option is set.
func (c *Client) ReverseGeocode(
	coord geo.Coordinate,
	opts ...geo.RGOption,
) ([]*geo.RGResult, error) {
	var options geo.RGOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Validate options.
	if options.IncludeShape && (options.Level == 0) {
		return nil, errors.Errorf("here: cannot include area shape without level " +
			"selection")
	}

	// Build request URL.
	url, err := url.Parse(reverseGeocodeURL)
	if err != nil {
		panic(err)
	}
	params := c.beginQuery(url)
	params.Set("gen", "9")
	params.Set("locationattributes", "address")

	if options.Level > 0 {
		level := options.Level.String()
		if options.Level == geo.PostcodeLevel {
			level = "postalCode"
		}
		params.Set("level", level)
		params.Set("mode", "retrieveAll")
	}
	{
		var radius uint = 50
		if options.Radius > 0 {
			radius = options.Radius
		}
		params.Set("prox", fmt.Sprintf("%f,%f,%d", coord.Y, coord.X, radius))
	}
	if options.IncludeShape {
		params.Set("additionalData", "IncludeShapeLevel,default")
	}

	url.RawQuery = params.Encode()

	res, err := c.httpc.Get(url.String())
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
						Shape *struct{ Value string }
					}
				}
			}
		}
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Errorf("here: decoding response body: %w", err)
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Errorf("here: closing response body: %w", err)
	}

	// Parse response.
	if len(data.Response.View) == 0 {
		return nil, errors.New("here: no result views")
	}

	var (
		matches = data.Response.View[0].Result
		results = make([]*geo.RGResult, len(matches))
	)
	for i, match := range matches {
		var (
			loc  = &match.Location
			pos  = &loc.Position
			addr = &loc.Address
		)

		var shape []geo.Coordinate
		if loc.Shape != nil {
			value := loc.Shape.Value
			if !strings.HasPrefix(value, "POLYGON") {
				goto EncodeResult
			}
			value = value[10 : len(value)-2]

			// Split value into coordinate pairs.
			pairs := strings.Split(value, ", ")
			for _, pair := range pairs {
				var (
					coords = strings.Fields(pair)
					fcs    = make([]float64, len(coords))
				)
				for j, coord := range coords {
					if fcs[j], err = strconv.ParseFloat(coord, 64); err != nil {
						return nil, errors.Errorf("here: parsing half-coordinate '%s': %w",
							coord, err)
					}
				}
				shape = append(shape, geo.Coordinate{X: fcs[0], Y: fcs[1]})
			}
		}

	EncodeResult:
		results[i] = &geo.RGResult{
			Location: geo.Location{
				ID:       loc.ID,
				Level:    match.MatchLevel,
				Type:     loc.Type,
				Position: geo.Coordinate{X: pos.Longitude, Y: pos.Latitude},
				Address: &geo.Address{
					Label:      addr.Label,
					Country:    addr.Country,
					State:      addr.State,
					County:     addr.County,
					City:       addr.City,
					District:   addr.District,
					PostalCode: addr.PostalCode,
					Street:     addr.Street,
					Number:     addr.HouseNumber,
				},
				Shape: shape,
			},
		}
	}
	return results, nil
}
