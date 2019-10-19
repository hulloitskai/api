package heretrans

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

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/assist/transit/transutil"
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
	coords location.Coordinates,
	cfg transit.NearbyDeparturesConfig,
) ([]transit.NearbyDeparture, error) {
	// Build and perform request.
	url := buildNearbyDeparturesURL(coords, &cfg)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "heretrans: create request")
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
							Time time.Time `json:"time"`
							RT   *struct {
								Dep time.Time `json:"dep"`
							}
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
		return nil, errors.Wrap(err, "heretrans: decode response body")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "heretrans: close response body")
	}

	// Marshal response to []transit.NearbyDeparture.
	var (
		ops = make(map[string]*transit.Operator)
		tps = make(map[uint32]*transit.Transport)
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
					ID:          s.ID,
					Name:        transutil.NormalizeStationName(s.Name),
					Coordinates: location.Coordinates{X: s.X, Y: s.Y},
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

			ndsByHash := make(map[uint32]*transit.NearbyDeparture)
			for j := range deps.Dep {
				var (
					dep  = &deps.Dep[j]
					dtp  = &dep.Transport
					hash = transutil.HashTransportComponents(
						dtp.Name, dtp.Dir, dtp.At.Operator,
					)
				)

				nd, ok := ndsByHash[hash]
				if !ok {
					tp, ok := tps[hash]
					if !ok {
						op, ok := ops[dtp.At.Operator]
						if !ok {
							return nil, errors.Newf(
								"heretrans: unknown operator with code '%s'",
								dtp.At.Operator,
							)
						}

						// Create transit.Transport.
						tp = &transit.Transport{
							Route:     dtp.Name,
							Direction: transutil.NormalizeStationName(dtp.Dir),
							Category:  dtp.At.Category,
							Operator:  op,
						}

						// Modify tp based on well-known op codes.
						switch tp.Operator.Code {
						case transit.OpCodeGoTransit:
							if tp.Category == "Bus" {
								if n := strings.IndexByte(tp.Direction, '-'); n != -1 {
									tp.Route = tp.Direction[:n-1]
									tp.Direction = tp.Direction[n+2:]
								}
							}
						case transit.OpCodeGRT:
							if tp.Category == "Bus" {
								if tp.Direction[1] == '-' {
									tp.Route += tp.Direction[:1]
									tp.Direction = tp.Direction[2:]
								}
							}
						}

						// Cache transit.Transport.
						tps[hash] = tp
					}

					// Construct transit.NearbyDeparture without times, and save to cache.
					nd = &transit.NearbyDeparture{
						Distance: distance,
						Departure: transit.Departure{
							Transport: tp,
							Station:   stn,
						},
					}
					ndsByHash[hash] = nd
				}

				if rt := dep.RT; rt != nil {
					// Only tag departure as realtime if first result is real-time.
					if len(nd.Times) == 0 {
						nd.Realtime = true
					}
					nd.Times = append(nd.Times, rt.Dep)
				} else {
					nd.Times = append(nd.Times, dep.Time)
				}
			}

			// Save ndsByHash to nds.
			for _, nd := range ndsByHash {
				nds = append(nds, *nd)
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

	if r := cfg.Radius; r != nil {
		ps.Set("radius", formatInt(*r))
	}
	if m := cfg.MaxPerStation; m != nil {
		ps.Set("max", formatInt(*m))
	}
	if m := cfg.MaxStations; m != nil {
		ps.Set("maxStn", formatInt(*m))
	}
	if m := cfg.MaxPerTransport; m != nil {
		ps.Set("maxPerTransport", formatInt(*m))
	}

	// Encode query params and return URL.
	u.RawQuery = ps.Encode()
	return u.String()
}

func formatInt(u int) string { return strconv.FormatInt(int64(u), 10) }
