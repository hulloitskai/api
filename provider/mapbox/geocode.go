package mapbox

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/stevenxie/api/pkg/geo"
	errors "golang.org/x/xerrors"
)

const geocodingURL = baseURL + "/geocoding/v5/mapbox.places"

// ReverseGeocode reverse-geocodes a coordinate.
func (c *Client) ReverseGeocode(
	coord geo.Coordinate,
	opts ...geo.GeocodeOption,
) ([]*geo.Feature, error) {
	var options geo.GeocodeOptions
	for _, opt := range opts {
		opt(&options)
	}

	// Build request URL.
	url, err := url.Parse(fmt.Sprintf("%s/%f,%f.json", geocodingURL, coord.X, coord.Y))
	if err != nil {
		return nil, errors.Errorf("mapbox: parsing URL: %w", err)
	}

	params := c.beginQuery(url)
	if len(options.Types) > 0 {
		types := make([]string, len(options.Types))
		for i, t := range options.Types {
			types[i] = string(t)
		}
		params.Set("types", strings.Join(types, ","))
	}
	url.RawQuery = params.Encode()

	// Perform request.
	res, err := c.httpc.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var data struct {
		Features []*geo.Feature `json:"features"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Errorf("mapbox: decoding response body: %w", err)
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Errorf("mapbox: closing response body: %w", err)
	}

	return data.Features, nil
}
