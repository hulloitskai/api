package heretrans

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/here"
)

// NewLocator creates a new transit.Locator.
func NewLocator(c here.Client) transit.Locator {
	return locator{
		client: c,
	}
}

type locator struct {
	client here.Client
}

var _ transit.Locator = (*locator)(nil)

func (l locator) NearbyDepartures(
	ctx context.Context,
	pos location.Coordinates,
	opts ...transit.NearbyDeparturesOption,
) ([]transit.NearbyDeparture, error) {
	var cfg transit.NearbyDeparturesConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	// Build and perform request.
	url := buildNearbyDeparturesURL(pos, &cfg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "heretrans: creating request")
	}
	res, err := l.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode response.
	var data struct {
		Res struct {
			MultiNextDepartures struct {
				MultiNextDeparture []struct {
					Stn struct {
						X        float64 `json:"x"`
						Y        float64 `json:"y"`
						ID       string  `json:"id"`
						Name     string  `json:"name"`
						Distance int     `json:"distance"`
					}
					NextDepartures struct {
						Dep []struct {
							Time      time.Time
							Transport struct {
								Dir  string `json:"dir"`
								Name string `json:"name"`
								At   struct {
									Category string `json:"category"`
									Operator string `json:"operator"`
								}
							}
						}
						Operators struct {
							Op []struct {
								Name string `json:"name"`
								Code string `json:"code"`
							}
						}
					}
				}
			}
		}
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Wrap(err, "heretrans: decoding response body")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "heretrans: closing response body")
	}

	// Marshal response to []transit.NearbyDeparture.
	var (
		ops = make(map[string]*transit.Operator)
		tps = make(map[string]*transit.Transport)
		nds []transit.NearbyDeparture
	)
	{
		sets := data.Res.MultiNextDepartures.MultiNextDeparture
		for i := range sets {
			set := &sets[i]

			// Build transit.Station.
			var (
				stn      *transit.Station
				distance int
			)
			{
				s := set.Stn
				stn = &transit.Station{
					ID:       s.ID,
					Name:     s.Name,
					Position: location.Coordinates{X: s.X, Y: s.Y},
				}
				distance = s.Distance
			}

			// Cache transit.Operators.
			deps := set.NextDepartures
			{
				for _, op := range deps.Operators.Op {
					if _, ok := ops[op.Code]; ok {
						continue
					}
					ops[op.Code] = &transit.Operator{
						Name: op.Name,
						Code: op.Code,
					}
				}
			}

			for j := range deps.Dep {
				dep := &deps.Dep[j]

				// Derive transport, add to cache if not yet exists.
				var tp *transit.Transport
				{
					var (
						dt  = &dep.Transport
						key = dt.Name + dt.Dir
						ok  bool
					)
					if tp, ok = tps[key]; !ok {
						op, ok := ops[dt.At.Operator]
						if !ok {
							return nil, errors.Newf(
								"heretrans: unknown operator with code '%s'",
								dt.At.Operator,
							)
						}
						tp = &transit.Transport{
							Route:     dt.Name,
							Direction: dt.Dir,
							Category:  dt.At.Category,
							Operator:  op,
						}
						tps[key] = tp
					}
				}

				// Build transit.NearbyDeparture.
				nds = append(nds, transit.NearbyDeparture{
					Distance: distance,
					Departure: transit.Departure{
						Times:     []time.Time{dep.Time},
						Transport: tp,
						Station:   stn,
					},
				})
			}
		}
	}
	return nds, nil
}

const _multiboardURL = "https://transit.api.here.com/v3/multiboard/" +
	"by_geocoord.json"

func buildNearbyDeparturesURL(
	pos location.Coordinates,
	cfg *transit.NearbyDeparturesConfig,
) string {
	u, err := url.Parse(_multiboardURL)
	if err != nil {
		panic(err)
	}

	// Set query parameters.
	ps := u.Query()
	ps.Set("time", time.Now().Format(time.RFC3339))
	ps.Set("center", fmt.Sprintf("%f,%f", pos.Y, pos.X))
	ps.Set("maxPerTransport", "1")

	if r := cfg.Radius; r != nil {
		ps.Set("radius", formatUint(*r))
	}
	if m := cfg.MaxPerStation; m != nil {
		ps.Set("max", formatUint(*m))
	}
	if m := cfg.MaxStations; m != nil {
		ps.Set("maxStn", formatUint(*m))
	}

	// Encode query params and return URL.
	u.RawQuery = ps.Encode()
	return u.String()
}

func formatUint(u uint) string { return strconv.FormatUint(uint64(u), 10) }
