package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
	"go.stevenxie.me/gopkg/zero"
)

// NewService creates a new auth.Service that is backed by Airtable.
func NewService(
	c Client,
	sel CodesSelector,
	opts ...ServiceOption,
) auth.Service {
	// Validate arguments.
	if c == nil {
		panic(errors.New("airtable: client may not be nil"))
	}

	// Configure and build service.
	cfg := ServiceConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &service{
		client: c,
		selectors: serviceSelectors{
			codes:  &sel,
			access: cfg.AccessSelector,
		},
		// timezone: cfg.Timezone,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type (
	service struct {
		client    Client
		selectors serviceSelectors

		// timezone *time.Location
		log *logrus.Entry
	}

	serviceSelectors struct {
		codes  *CodesSelector
		access *AccessSelector
	}

	// A ServiceConfig configures a auth.Service.
	ServiceConfig struct {
		// If provided, permissions access records will be saved in the
		// corresponding fields.
		AccessSelector *AccessSelector

		// Timezone *time.Location
		Logger *logrus.Entry
	}

	// A ServiceOption modifies a ServiceConfig.
	ServiceOption func(*ServiceConfig)
)

var _ auth.Service = (*service)(nil)

func (svc *service) GetPermissions(
	ctx context.Context,
	code string,
) ([]auth.Permission, error) {
	perms, _, err := svc.getPermissions(ctx, code)
	return perms, err
}

func (svc *service) HasPermission(
	ctx context.Context,
	code string, p auth.Permission,
) (ok bool, err error) {
	// Validate inputs.
	if p == "" {
		return false, errors.Newf("airtable: invalid permission")
	}

	perms, id, err := svc.getPermissions(ctx, code)
	if err != nil {
		if !errors.Is(err, auth.ErrInvalidCode) {
			err = errors.Wrap(err, "airtable: getting permissions")
		}
		return false, err
	}
	for _, perm := range perms {
		if p == perm {
			go svc.recordAccess(id, p)
			return true, nil
		}
	}
	return false, nil
}

func (svc *service) getPermissions(
	ctx context.Context,
	code string,
) (perms []auth.Permission, recordID string, err error) {
	// Validate inputs.
	if code == "" {
		return nil, "", auth.ErrInvalidCode
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(svc.getPermissions),
		"code":            code,
	}).WithContext(ctx)

	var (
		selector = svc.selectors.codes
		offset   *string // used for pagination.
	)
	{
		empty := ""
		offset = &empty
	}
	log = log.WithField("selector", selector)

	// Start building request URL.
	url, err := selector.BuildURL()
	if err != nil {
		log.WithError(err).Error("Failed to build selection URL.")
		return nil, "", errors.Wrap(err, "airtable: building selection URL")
	}
	params := url.Query()

	for offset != nil {
		log := log
		if *offset != "" {
			log = log.WithField("offset", *offset)

			// Set query params.
			params.Set("offset", *offset)
			url.RawQuery = params.Encode()
		}

		// Perform request.
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodGet, url.String(),
			nil,
		)
		if err != nil {
			log.WithError(err).Error("Failed to create request.")
			return nil, "", errors.Wrap(err, "airtable: create request")
		}
		res, err := svc.client.Do(req)
		if err != nil {
			log.WithError(err).Error("Failed to retrieve records.")
			return nil, "", err
		}
		defer res.Body.Close()

		// Decode response as JSON.
		var data struct {
			Records []struct {
				ID     string                    `json:"id"`
				Fields map[string]zero.Interface `json:"fields"`
			} `json:"records"`
			Offset *string `json:"offset"`
		}
		if err = json.NewDecoder(res.Body).Decode(&data); err != nil {
			log.WithError(err).Error("Failed to decode response as JSON.")
			return nil, "", errors.Wrap(err, "airtable: decoding response as JSON")
		}
		if err = res.Body.Close(); err != nil {
			log.WithError(err).Error("Failed to close response body.")
			return nil, "", errors.Wrap(err, "airtable: closing response body")
		}

		// Range through each record; if the codes match, return permissions.
		for _, record := range data.Records {
			log := log.WithField("record_id", record.ID)
			var fields struct {
				Code  string   `json:"code"`
				Perms []string `json:"auth"`
			}
			if err := mapstructure.Decode(record.Fields, &fields); err != nil {
				log.WithError(err).Error("Failed to parse record fields.")
				return nil, "", errors.Wrapf(
					err,
					"airtable: parsing fields for record with ID '%s'", record.ID,
				)
			}

			// Check to see if the code matches.
			if c := strings.TrimSpace(fields.Code); c != code {
				continue
			}

			// Marshal to auth.Permissions.
			permissions := make([]auth.Permission, len(fields.Perms))
			for i, p := range fields.Perms {
				permissions[i] = auth.Permission(p)
			}
			return permissions, record.ID, nil
		}

		offset = data.Offset
	}

	// All records have been checked; no such code.
	return nil, "", auth.ErrInvalidCode
}

func (svc service) recordAccess(id string, p auth.Permission) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.recordAccess),
		"id":              id,
	})

	// If no access selector, skip.
	selector := svc.selectors.access
	if selector == nil {
		log.Warn("No access mapping; skipping.")
		return
	}
	log.WithField("selector", selector)

	// Map access record to table fields.
	fields := map[string]zero.Interface{
		selector.FieldSelector.Time:         time.Now().Format(time.RFC3339),
		selector.FieldSelector.Perm:         string(p),
		selector.FieldSelector.CodeRecordID: []string{id},
	}

	type recordData struct {
		Fields map[string]zero.Interface `json:"fields"`
	}
	data := struct {
		Records []recordData `json:"records"`
	}{[]recordData{recordData{Fields: fields}}}

	// Encode fields as JSON.
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&data); err != nil {
		log.WithError(err).Error("Failed to encode fields as JSON.")
		return
	}

	// Build request.
	url, err := selector.BuildURL()
	if err != nil {
		log.WithError(err).Error("Failed to build URL.")
		return
	}
	req, err := http.NewRequest(http.MethodPost, url.String(), buf)
	if err != nil {
		log.WithError(err).Error("Failed to create request.")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform request.
	res, err := svc.client.Do(req)
	if err != nil {
		log.WithError(err).Error("Failed to send access update.")
		return
	}
	if res.StatusCode != http.StatusOK {
		entry := log.WithField("status", res.StatusCode)
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			entry = log.WithField("body", string(body))
		}
		entry.Error("Bad status in access update response.")
		return
	}
}
