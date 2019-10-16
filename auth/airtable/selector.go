package airtable

import (
	"fmt"
	"net/url"

	"github.com/cockroachdb/errors"
	validation "github.com/go-ozzo/ozzo-validation"
)

const _baseURL = "https://api.airtable.com/v0"

// DefaultCodesSelector creates a CodesSelector with default values.
func DefaultCodesSelector() CodesSelector {
	var sel CodesSelector
	sel.FieldSelector.Code = "code"
	sel.FieldSelector.Perms = "perms"
	return sel
}

// DefaultAccessSelector creates a AccessSelector with default values.
func DefaultAccessSelector() AccessSelector {
	var sel AccessSelector
	sel.FieldSelector.Time = "time"
	sel.FieldSelector.Perm = "perm"
	sel.FieldSelector.CodeRecordID = "code-record-id"
	return sel
}

type (
	// A CodesSelector locates the fields within an AirTable base that are used
	// to store auth codes.
	CodesSelector struct {
		ViewSelector  `yaml:"viewSelector"`
		FieldSelector struct {
			Code  string `yaml:"code"`
			Perms string `yaml:"perms"`
		} `yaml:"fieldSelector"`
	}

	// An AccessSelector locates the fields within an AirTable base that are used
	// to store data-access attempts.
	AccessSelector struct {
		ViewSelector  `yaml:"viewSelector"`
		FieldSelector struct {
			Time         string `yaml:"time"`
			Perm         string `yaml:"perm"`
			CodeRecordID string `yaml:"codeRecordID"`
		} `yaml:"fieldSelector"`
	}
)

var (
	_ validation.Validatable = (*CodesSelector)(nil)
	_ validation.Validatable = (*AccessSelector)(nil)
)

// Validate returns an error if the PermsSelector is not valid.
func (sel *CodesSelector) Validate() error {
	if err := validation.Validate(&sel.ViewSelector); err != nil {
		return errors.Wrap(err, "validating ViewSelector")
	}
	{
		fields := &sel.FieldSelector
		if err := validation.ValidateStruct(
			fields,
			validation.Field(&fields.Code, validation.Required),
			validation.Field(&fields.Perms, validation.Required),
		); err != nil {
			return errors.Wrap(err, "validating FieldSelector")
		}
	}
	return nil
}

// Validate returns an error if the AccessSelector is not valid.
func (sel *AccessSelector) Validate() error {
	if err := validation.Validate(&sel.ViewSelector); err != nil {
		return errors.Wrap(err, "validating ViewSelector")
	}
	{
		fields := &sel.FieldSelector
		if err := validation.ValidateStruct(
			fields,
			validation.Field(&fields.Time, validation.Required),
			validation.Field(&fields.CodeRecordID, validation.Required),
		); err != nil {
			return errors.Wrap(err, "validating FieldSelector")
		}
	}
	return nil
}

// A ViewSelector selects an Airtable view.
type ViewSelector struct {
	BaseID    string `yaml:"baseID"`
	TableName string `yaml:"tableName"`
	ViewName  string `yaml:"viewName"`
}

var _ validation.Validatable = (*ViewSelector)(nil)

// BuildURL builds the URL for the corresponding view.
func (sel *ViewSelector) BuildURL() (*url.URL, error) {
	raw := fmt.Sprintf("%s/%s/%s", _baseURL, sel.BaseID, sel.TableName)
	url, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}

	if sel.ViewName != "" {
		// Encode view as a query parameter.
		ps := url.Query()
		ps.Set("view", sel.ViewName)
		url.RawQuery = ps.Encode()
	}
	return url, nil
}

// Validate returns an error if the ViewSelector is not valid.
func (sel *ViewSelector) Validate() error {
	var (
		reqs  = []*string{&sel.BaseID, &sel.TableName}
		rules = make([]*validation.FieldRules, len(reqs))
	)
	for i, req := range reqs {
		rules[i] = validation.Field(req, validation.Required)
	}
	return validation.ValidateStruct(sel, rules...)
}
