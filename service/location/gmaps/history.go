package gmaps

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/service/location"
)

const kmlURL = "https://www.google.com/maps/timeline/kml"

// LocationHistory looks up the authenticated user's location history for a
// particular date.
func (h *Historian) LocationHistory(date time.Time) (location.HistorySegments,
	error) {
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
		return nil, errors.Newf("maps: bad response status '%d'", res.StatusCode)
	}
	defer res.Body.Close()

	// Decode response as XML.
	var data struct {
		Placemarks []struct {
			Name        string `xml:"name"`
			Address     string `xml:"address"`
			Description string `xml:"description"`
			TimeSpan    struct {
				Begin time.Time `xml:"begin"`
				End   time.Time `xml:"end"`
			}
			Data []struct {
				Name  string `xml:"name,attr"`
				Value string `xml:"value"`
			} `xml:"ExtendedData>Data"`
			LineStringCoordinates string `xml:"LineString>coordinates"`
			PointCoordinates      string `xml:"Point>coordinates"`
		} `xml:"Document>Placemark"`
	}
	if err = xml.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "maps: decoding response as XML")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "maps: closing response body")
	}

	results := make(location.HistorySegments, len(data.Placemarks))
	for i, pm := range data.Placemarks {
		var (
			span    = &pm.TimeSpan
			segment = &location.HistorySegment{
				Place:       pm.Name,
				Address:     pm.Address,
				Description: strings.TrimSpace(pm.Description),
				TimeSpan: location.TimeSpan{
					Begin: span.Begin,
					End:   span.End,
				},
			}
		)

		// Parse metadata.
		for _, data := range pm.Data {
			switch data.Name {
			case "Category":
				segment.Category = data.Value
			case "Distance":
				if segment.Distance, err = strconv.Atoi(data.Value); err != nil {
					return nil, errors.Wrap(err, "maps: parsing distance as int")
				}
			}
		}

		// Parse coordinates.
		for _, raw := range []string{
			pm.LineStringCoordinates,
			pm.PointCoordinates,
		} {
			if raw == "" {
				continue
			}
			var (
				rawcoords = strings.Fields(raw)
				coords    = make([]location.Coordinates, len(rawcoords))
			)
			for j, triplet := range rawcoords {
				var (
					unitStrings = strings.Split(triplet, ",")
					units       = make([]float64, len(unitStrings))
				)
				for k, us := range unitStrings {
					if units[k], err = strconv.ParseFloat(us, 64); err != nil {
						return nil, errors.Wrap(err, "maps: parsing coordinate unit as int")
					}
				}
				coords[j] = location.Coordinates{
					X: units[0],
					Y: units[1],
					Z: units[2],
				}
			}
			segment.Coordinates = coords
		}

		results[i] = segment
	}

	return results, nil
}

// RecentHistory returns the authenticated user's recent location history.
func (h *Historian) RecentHistory() (location.HistorySegments, error) {
	date := time.Now().In(h.timezone)
	segments, err := h.LocationHistory(date)
	if err != nil {
		return nil, err
	}
	if len(segments) > 0 {
		return segments, nil
	}

	// Fallback to previous date if no history is recorded.
	return h.LocationHistory(date.Add(-24 * time.Hour))
}
