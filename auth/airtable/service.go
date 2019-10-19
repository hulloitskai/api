package airtable

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/openlyinc/pointy"

	"github.com/cockroachdb/errors"
	"github.com/mitchellh/mapstructure"
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
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*service).HasPermission),
		"code":            code,
		"permission":      p,
	}).WithContext(ctx)

	// Validate inputs.
	if p == "" {
		log.Warn("Invalid permission (empty).")
		return false, errors.Newf("airtable: invalid permission")
	}

	log.Trace("Getting permissions for code...")
	perms, id, err := svc.getPermissions(ctx, code)
	if err != nil {
		if !errors.Is(err, auth.ErrInvalidCode) {
			log.WithError(err).Error("Failed to get code permissions.")
			err = errors.Wrap(err, "airtable: getting code permissions")
		}
		return false, err
	}
	log = log.WithField("code_permissions", perms)
	log.Trace("Got code permissions.")

	for _, perm := range perms {
		if p == perm {
			log.Trace("Requested permission matches code permissions.")
			go svc.recordAccess(id, p)
			return true, nil
		}
	}

	log.Trace("No matching code permission.")
	return false, nil
}

func (svc *service) getPermissions(
	ctx context.Context,
	code string,
) (perms []auth.Permission, recordID string, err error) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod((*service).getPermissions),
		"code":            code,
	}).WithContext(ctx)

	// Validate inputs.
	if code == "" {
		log.Warn("Code is invalid (empty).")
		return nil, "", auth.ErrInvalidCode
	}

	var (
		selector = svc.selectors.codes
		offset   = pointy.String("") // used for pagination.
	)
	log = log.WithField("selector", selector)

	// Build initial request URL.
	u, err := selector.BuildURL()
	if err != nil {
		log.WithError(err).Error("Failed to build selection URL.")
		return nil, "", errors.Wrap(err, "airtable: building selection URL")
	}
	params := u.Query()

	for offset != nil {
		log := log
		if *offset != "" {
			log = log.WithField("offset", *offset)
		}

		// Build final URL.
		params.Set("offset", *offset)
		u.RawQuery = params.Encode()
		url := u.String()

		// Perform request.
		log.WithField("url", url).Trace("Getting records matching selection...")
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
		log.WithField("response", data).Trace("Got response data.")

		// Range through each record; if the codes match, return permissions.
		log.Trace("Checking records for matching code.")
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
			log.WithField("fields", fields).Trace("Decoded record fields.")

			// Check to see if the code matches.
			if c := strings.TrimSpace(fields.Code); c != code {
				continue
			}
			log.Trace("Found field with matching code.")

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
	log.Warn("No matching records found; invalid code.")
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
	log = log.WithField("selector", selector)
	log.Trace("Recording permissions record access...")

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
	log.WithField("payload", data).Trace("Constructed request payload.")

	// Encode fields as JSON.
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(&data); err != nil {
		log.WithError(err).Error("Failed to encode fields as JSON.")
		return
	}

	// Build request.
	u, err := selector.BuildURL()
	if err != nil {
		log.WithError(err).Error("Failed to build URL.")
		return
	}
	url := u.String()

	log.WithField("url", url).Trace("Sending record update request...")
	req, err := http.NewRequest(http.MethodPost, url, buf)
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
		log := log.WithField("status", res.StatusCode)
		if body, err := ioutil.ReadAll(res.Body); err == nil {
			log = log.WithField("response", string(body))
		}
		log.Error("Bad status in access update response.")
		return
	}
	log.Trace("Record update success!")
}
