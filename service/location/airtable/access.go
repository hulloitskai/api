package airtable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/pkg/zero"
	"go.stevenxie.me/api/service/location"
)

type (
	// A LocationAccessService implements a location.AccessService using Airtable.
	LocationAccessService struct {
		client *Client
		baseID string
		table  string
		view   string

		timezone *time.Location
		log      logrus.FieldLogger
	}

	// A LocationAccessServiceConfig configures a LocationAccessService.
	LocationAccessServiceConfig struct {
		Timezone *time.Location
		Logger   logrus.FieldLogger
	}
)

var _ location.AccessService = (*LocationAccessService)(nil)

// NewLocationAccessService creates a new LocationAccessService that checks
// Airtable for records of non-expired access codes.
func NewLocationAccessService(
	c *Client,
	baseID, table, view string,
	opts ...func(*LocationAccessServiceConfig),
) *LocationAccessService {
	cfg := LocationAccessServiceConfig{
		Logger: zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &LocationAccessService{
		client: c,
		baseID: baseID,
		table:  table,
		view:   view,

		timezone: cfg.Timezone,
		log:      cfg.Logger,
	}
}

// IsValidCode returns true if code is a valid location access code, and false
// otherwise.
func (svc *LocationAccessService) IsValidCode(code string) (bool, error) {
	var offset *string

Fetch:
	// Create and send request.
	url, err := url.Parse(svc.tableURL())
	if err != nil {
		return false, errors.Wrap(err, "airtable: building request URL")
	}
	{
		ps := url.Query()
		ps.Set("view", svc.view)
		if offset != nil {
			ps.Set("offset", *offset)
		}
		if svc.timezone != nil {
			ps.Set("timezone", svc.timezone.String())
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
				Code     string `json:"code"`
				Accesses int    `json:"accesses"`
			} `json:"fields"`
		} `json:"records"`
		Offset *string `json:"offset"`
	}
	if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
		return false, errors.Wrap(err, "airtable: decoding response as JSON")
	}
	if err = res.Body.Close(); err != nil {
		return false, errors.Wrap(err, "airtable: closing response body")
	}

	// Determine of code is included in data.Records.
	for _, record := range data.Records {
		fields := &record.Fields
		if fields.Code == code {
			go svc.recordAccess(record.ID, fields.Accesses+1) // async update
			return true, nil
		}
	}
	if data.Offset != nil { // try next set of records
		offset = data.Offset
		goto Fetch
	}
	return false, nil
}

func (svc *LocationAccessService) tableURL() string {
	return fmt.Sprintf("%s/%s/%s", baseURL, svc.baseID, svc.table)
}

func (svc *LocationAccessService) recordAccess(id string, accesses int) {
	entry := svc.log.WithField("id", id)

	var data struct {
		Fields struct {
			Accesses     int    `json:"accesses"`
			LastAccessed string `json:"last-accessed"`
		} `json:"fields"`
	}
	data.Fields.Accesses = accesses
	data.Fields.LastAccessed = time.Now().In(svc.timezone).Format(time.RFC3339)

	// Encode data as JSON.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		entry.WithError(err).Error("Failed to encode access update as JSON.")
	}

	var (
		url      = fmt.Sprintf("%s/%s", svc.tableURL(), id)
		req, err = http.NewRequest(http.MethodPatch, url, &buf)
	)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := svc.client.Do(req)
	if err != nil {
		entry.WithError(err).Error("Failed to send access update.")
	}
	if res.StatusCode != http.StatusOK {
		entry.
			WithField("status", res.StatusCode).
			Error("Bad status in access update response.")
	}
}
