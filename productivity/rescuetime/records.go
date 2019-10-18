package rescuetime

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/productivity"
)

const (
	_baseURL = "https://www.rescuetime.com/anapi/data"
	_iso8601 = "2006-01-02"
)

// NewRecordSource creates a new productivity.RecordSource.
func NewRecordSource(c Client) productivity.RecordSource {
	return recordSource{
		client: c,
	}
}

type recordSource struct {
	client Client
	loc    *time.Location
}

var _ productivity.RecordSource = (*recordSource)(nil)

// CurrentProductivity gets the current productivity.Periods using data from
// RescueTime.
func (svc recordSource) GetRecords(
	ctx context.Context,
	date time.Time,
) ([]productivity.Record, error) {
	ds := date.Format(_iso8601)

	// Build query params.
	params := make(url.Values)
	params.Set("version", "0")
	params.Set("format", "json")
	params.Set("restrict_begin", ds)
	params.Set("restrict_end", ds)
	params.Set("restrict_kind", "productivity")

	// Send request.
	u, err := url.Parse(_baseURL)
	if err != nil {
		panic(err)
	}
	u.RawQuery = params.Encode()

	// Perform request.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "rescuetime: create request")
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
	records := make([]productivity.Record, len(data.Rows))
	for i, row := range data.Rows {
		records[i] = productivity.Record{
			Category: productivity.Category(row[3] + 3),
			Duration: time.Duration(row[1]) * time.Second,
		}
	}
	return records, nil
}
