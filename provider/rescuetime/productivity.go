package rescuetime

import (
	"encoding/json"
	"net/url"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/api"
)

const (
	baseURL = "https://www.rescuetime.com/anapi/data"
	iso8601 = "2006-01-02"
)

// CurrentProductivity gets the current api.Productivity values from
// RescueTime.
func (c *Client) CurrentProductivity() ([]*api.ProductivitySegment, error) {
	nowstr := c.currentDate()

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
		return nil, errors.Wrap(err, "rescuetime: decoding response as JSON")
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Wrap(err, "rescuetime: closing response body")
	}

	// Parse productivity data.
	segs := make([]*api.ProductivitySegment, len(data.Rows))
	for i, row := range data.Rows {
		var name string
		switch row[3] {
		case 2:
			name = "Very Productive"
		case 1:
			name = "Productive"
		case 0:
			name = "Neutral"
		case -1:
			name = "Distracting"
		case -2:
			name = "Very Distracting"
		default:
			return nil, errors.Newf(
				"rescuetime: unknown productivity ID '%d'",
				row[3],
			)
		}

		segs[i] = &api.ProductivitySegment{
			Name:     name,
			ID:       row[3],
			Duration: row[1],
		}
	}

	// Sort segments by ID.
	sort.Sort(sortableSegs(segs))
	return segs, nil
}

func (c *Client) currentDate() string {
	now := time.Now()
	if c.timezone != nil {
		now = now.In(c.timezone)
	}
	return now.Format(iso8601)
}

// queryParams returns a set of default query params to send with requests
// to the RescueTime API.
func (c *Client) queryParams() url.Values {
	qp := make(url.Values)
	qp.Set("key", c.key)
	qp.Set("version", "0")
	return qp
}

// sortableSegs implements sort.Interface for a slice of
// api.ProductivitySegments.
type sortableSegs []*api.ProductivitySegment

var _ sort.Interface = (*sortableSegs)(nil)

func (segs sortableSegs) Len() int           { return len(segs) }
func (segs sortableSegs) Less(i, j int) bool { return segs[i].ID < segs[j].ID }
func (segs sortableSegs) Swap(i, j int) {
	segs[i], segs[j] = segs[j], segs[i]
}
