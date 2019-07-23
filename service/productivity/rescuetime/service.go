package rescuetime

import (
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/service/productivity"
)

const (
	baseURL = "https://www.rescuetime.com/anapi/data"
	iso8601 = "2006-01-02"
)

type (
	// A ProductivityService implements a productivity.Service using a RescueTime
	// Client.
	ProductivityService struct {
		client   *Client
		timezone *time.Location
	}

	// ProductivityServiceConfig configures a ProductivityService.
	ProductivityServiceConfig struct {
		Timezone *time.Location
	}
)

var _ productivity.Service = (*ProductivityService)(nil)

// NewProductivityService creates a new ProductivityService.
func NewProductivityService(
	c *Client,
	opts ...func(*ProductivityServiceConfig),
) *ProductivityService {
	var cfg ProductivityServiceConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	return &ProductivityService{
		client:   c,
		timezone: cfg.Timezone,
	}
}

// CurrentProductivity gets the current productivity.Periods using data from
// RescueTime.
func (svc *ProductivityService) CurrentProductivity() (productivity.Periods,
	error) {
	now := svc.currentDate()

	// Build query params.
	params := make(url.Values)
	params.Set("version", "0")
	params.Set("format", "json")
	params.Set("restrict_begin", now)
	params.Set("restrict_end", now)
	params.Set("restrict_kind", "productivity")

	// Send request.
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	u.RawQuery = params.Encode()

	// Create and perform request.
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "rescuetime: creating request")
	}
	res, err := svc.client.Do(req)
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
	periods := make(productivity.Periods, len(data.Rows))
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

		periods[i] = &productivity.Period{
			Name:     name,
			ID:       row[3],
			Duration: row[1],
		}
	}

	// Sort periods by ID.
	sort.Sort(sortablePeriods(periods))
	return periods, nil
}

func (svc *ProductivityService) currentDate() string {
	now := time.Now()
	if svc.timezone != nil {
		now = now.In(svc.timezone)
	}
	return now.Format(iso8601)
}

// sortablePeriods implements sort.Interface for productivity.Periods.
type sortablePeriods productivity.Periods

var _ sort.Interface = (*sortablePeriods)(nil)

func (segs sortablePeriods) Len() int {
	return len(segs)
}
func (segs sortablePeriods) Less(i, j int) bool {
	return segs[i].ID < segs[j].ID
}
func (segs sortablePeriods) Swap(i, j int) {
	segs[i], segs[j] = segs[j], segs[i]
}
