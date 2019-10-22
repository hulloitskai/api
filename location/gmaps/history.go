package gmaps

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/scheduling"
)

// NewHistorian creates a new location.Historian that can load
// location history from Google Maps.
func NewHistorian(client TimelineClient) location.Historian {
	return historian{
		client: client,
	}
}

type historian struct {
	client TimelineClient
}

var _ location.Historian = (*historian)(nil)

const _kmlURL = "https://www.google.com/maps/timeline/kml"

func (svc historian) GetHistory(
	ctx context.Context,
	date time.Time,
) ([]location.HistorySegment, error) {
	// Derive URL.
	year, month, day := date.Date()
	month--
	url := fmt.Sprintf(
		"%s?pb=!1m8!1m3!1i%d!2i%d!3i%d!2m3!1i%d!2i%d!3i%d", _kmlURL,
		year, month, day,
		year, month, day,
	)

	// Perform request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "gmaps: create request")
	}
	res, err := svc.client.Do(req)
	if err != nil {
		return nil, nil
	}
	if res.StatusCode != http.StatusOK {
		return nil, errors.Newf("gmaps: bad response status '%d'", res.StatusCode)
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
		return nil, errors.Wrap(err, "gmaps: decode response as XML")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "gmaps: close response body")
	}

	results := make([]location.HistorySegment, len(data.Placemarks))
	for i, pm := range data.Placemarks {
		var (
			span = &pm.TimeSpan
			seg  = &results[i]
		)
		*seg = location.HistorySegment{
			Place:       pm.Name,
			Address:     pm.Address,
			Description: strings.TrimSpace(pm.Description),
			TimeSpan: scheduling.TimeSpan{
				Start: span.Begin,
				End:   span.End,
			},
		}

		// Parse metadata.
		for _, data := range pm.Data {
			switch data.Name {
			case "Category":
				seg.Category = data.Value
			case "Distance":
				if seg.Distance, err = strconv.Atoi(data.Value); err != nil {
					return nil, errors.Wrap(err, "gmaps: parsing distance as int")
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
				rawCoords = strings.Fields(raw)
				coords    = &seg.Coordinates
			)
			*coords = make([]location.Coordinates, len(rawCoords))
			for j, triplet := range rawCoords {
				var (
					unitStrings = strings.Split(triplet, ",")
					units       = make([]float64, len(unitStrings))
				)
				for k, us := range unitStrings {
					if units[k], err = strconv.ParseFloat(us, 64); err != nil {
						return nil, errors.Wrap(err, "gmaps: parsing coordinate unit as int")
					}
				}
				(*coords)[j] = location.Coordinates{
					X: units[0],
					Y: units[1],
					Z: units[2],
				}
			}
		}
	}

	return results, nil
}
