package rescuetime

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/stevenxie/api/pkg/metrics"
	errors "golang.org/x/xerrors"
)

const baseURL = "https://www.rescuetime.com/anapi/data"

// CurrentProductivity gets the current metrics.Productivity values from
// RescueTime.
func (c *Client) CurrentProductivity() (*metrics.Productivity, error) {
	nowstr := time.Now().Format("2006-01-02")

	// Build query params.
	qp := c.queryParams()
	qp.Set("format", "json")
	qp.Set("restrict_begin", nowstr)
	qp.Set("restrict_end", nowstr)
	qp.Set("restrict_kind", "productivity")

	// Send request.
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	u.RawQuery = qp.Encode()
	res, err := c.httpc.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Decode response as JSON.
	var data struct {
		Rows [][]int `json:"rows"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, errors.Errorf("rescuetime: decoding response as JSON: %w", err)
	}

	// Parse productivity data.
	prod := new(metrics.Productivity)
	for _, row := range data.Rows {
		val := row[1]
		switch row[3] {
		case 2:
			prod.VeryProductive = val
		case 1:
			prod.Productive = val
		case 0:
			prod.Neutral = val
		case -1:
			prod.Distracting = val
		case -2:
			prod.VeryDistracting = val
		default:
			return nil, errors.Errorf("rescuetime: unknown productivity ID '%d'",
				row[3])
		}
	}
	return prod, nil
}

// queryParams returns a set of default query params to send with requests
// to the RescueTime API.
func (c *Client) queryParams() url.Values {
	qp := make(url.Values)
	qp.Set("key", c.key)
	qp.Set("version", "0")
	return qp
}
