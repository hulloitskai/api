package maps

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/provider/google"
)

type (
	// A Historian can load the authenticated user's location history from
	// Google Maps.
	Historian struct {
		client *http.Client
	}

	// A HOption configures a Historian.
	HOption func(h *Historian)
)

// NewHistorian creates a new Historian.
//
// If it is not created with a custom http.Client, it will attempt to create
// one using the following envvars:
//   - GOOGLE_HSID
//   - GOOGLE_SID
//   - GOOGLE_SSID
//
// These environment variables can be gleaned from the application cookies
// set by Google when you log in to Google Maps (check your web console for the
// cookies named 'SID', 'HSID', and 'SSID').
func NewHistorian(opts ...HOption) (*Historian, error) {
	h := new(Historian)
	for _, opt := range opts {
		opt(h)
	}

	if h.client == nil { // derive authenticated client from envvars
		var (
			vars   = []string{"HSID", "SID", "SSID"}
			varmap = make(map[string]string, len(vars))
		)
		for _, v := range vars {
			var (
				name    = fmt.Sprintf("%s_%s", strings.ToUpper(google.Namespace), v)
				val, ok = os.LookupEnv(name)
			)
			if !ok {
				return nil, errors.Errorf("maps: no such envvar '%s'", name)
			}
			varmap[v] = val
		}

		jar, err := cookiesFromMap(varmap)
		if err != nil {
			return nil, errors.Errorf("maps: constructing cookies: %w", err)
		}
		h.client = &http.Client{Jar: jar}
	}

	return h, nil
}

// WithHClient configures a Historian to make HTTP requests with c.
func WithHClient(c *http.Client) HOption {
	return func(h *Historian) { h.client = c }
}

const kmlURL = "https://www.google.com/maps/timeline/kml"

// LocationHistory looks up the authenticated user's location history for a
// particular date.
func (h *Historian) LocationHistory(date time.Time) ([]*geo.Placemark, error) {
	year, month, day := date.Date()
	month--
	url := fmt.Sprintf(
		"%s?pb=!1m8!1m3!1i%d!2i%d!3i%d!2m3!1i%d!2i%d!3i%d", kmlURL,
		year, month, day,
		year, month, day,
	)

	res, err := h.client.Get(url)
	if err != nil {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("maps: bad response status '%d'", res.StatusCode)
	}
	defer res.Body.Close()

	// Decode response as XML.
	var data struct {
		Placemarks []struct {
			*geo.Placemark
			Data []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value"`
			} `xml:"ExtendedData>Data"`
			LineStringCoordinates string `xml:"LineString>coordinates"`
			PointCoordinates      string `xml:"Point>coordinates"`
		} `xml:"Document>Placemark"`
	}
	if err = xml.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Errorf("maps: decoding response as XML: %w", err)
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Errorf("maps: closing response body: %w", err)
	}

	results := make([]*geo.Placemark, len(data.Placemarks))
	for i, placemark := range data.Placemarks {
		placemark.Description = strings.TrimSpace(placemark.Description)

		// Parse metadata.
		for _, data := range placemark.Data {
			switch data.Name {
			case "Category":
				placemark.Category = data.Value
			case "Distance":
				if placemark.Distance, err = strconv.Atoi(data.Value); err != nil {
					return nil, errors.Errorf("maps: parsing distance as int: %w", err)
				}
			}
		}

		// Parse coordinates.
		for _, raw := range []string{
			placemark.LineStringCoordinates,
			placemark.PointCoordinates,
		} {
			if raw == "" {
				continue
			}
			var (
				rawcoords = strings.Fields(raw)
				coords    = make([]geo.Coordinate, len(rawcoords))
			)
			for j, triplet := range rawcoords {
				var (
					unitStrings = strings.Split(triplet, ",")
					units       = make([]float64, len(unitStrings))
				)
				for k, us := range unitStrings {
					if units[k], err = strconv.ParseFloat(us, 64); err != nil {
						return nil, errors.Errorf("maps: parsing coordinate unit as "+
							"int: %w", err)
					}
				}
				coords[j] = geo.Coordinate{X: units[0], Y: units[1], Z: units[2]}
			}
			placemark.Coordinates = coords
		}

		results[i] = placemark.Placemark
	}

	return results, nil
}
