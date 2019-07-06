package airtable

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/stevenxie/api/pkg/api"

	errors "golang.org/x/xerrors"
)

// A LocationAccessService can validate location access codes.
type LocationAccessService struct {
	client *Client
	baseID string
	table  string
	view   string
}

var _ api.LocationAccessService = (*LocationAccessService)(nil)

// NewLocationAccessService creates a new LocationAccessService that checks
// Airtable for records of non-expired access codes.
func NewLocationAccessService(
	c *Client,
	baseID, table, view string,
) *LocationAccessService {
	return &LocationAccessService{
		client: c,
		baseID: baseID,
		table:  table,
		view:   view,
	}
}

// IsValidCode returns true if code is a valid location access code, and false
// otherwise.
func (svc *LocationAccessService) IsValidCode(code string) (bool, error) {
	var offset *string

Fetch:
	// Create and send request.
	url, err := url.Parse(fmt.Sprintf("%s/%s/location", baseURL, svc.baseID))
	if err != nil {
		panic(err)
	}
	{
		ps := url.Query()
		ps.Set("view", svc.view)
		if offset != nil {
			ps.Set("offset", *offset)
		}
		url.RawQuery = ps.Encode()
	}

	res, err := svc.client.Get(url.String())
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	// Decode response as JSON.
	var data struct {
		Records []struct {
			ID     string `json:"id"`
			Fields struct {
				Code string `json:"code"`
			} `json:"fields"`
		} `json:"records"`
		Offset *string `json:"offset"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return false, errors.Errorf("airtable: decoding response as JSON: %w", err)
	}
	if err = res.Body.Close(); err != nil {
		return false, errors.Errorf("airtable: closing response body: %w", err)
	}

	// Determine of code is included in data.Records.
	for _, record := range data.Records {
		if record.Fields.Code == code {
			return true, nil
		}
	}
	if data.Offset != nil { // try next set of records
		offset = data.Offset
		goto Fetch
	}
	return false, nil
}
