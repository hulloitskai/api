package airtable

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	ms "github.com/mitchellh/mapstructure"
	ess "github.com/unixpickle/essentials"
)

// BaseURL is the base url for the Airtable API.
const BaseURL = "https://api.airtable.com/v0"

type fetchOpts struct {
	Limit int
	Sort  []sortConfig
}

type sortConfig struct {
	Field     string `mapstructure:"field"`
	Direction string `mapstructure:"direction"`
}

// fetchRecords fetches `limit` records from `table` in Airtable, and unmarshals
// them into v.
func (c *Client) fetchRecords(table, v interface{}, opts *fetchOpts) error {
	// Construct and perform request.
	u, err := url.Parse(fmt.Sprintf("%s/%s/%s", BaseURL, c.cfg.BaseID, table))
	if err != nil {
		panic(err)
	}

	// Configure request according to opts.
	if opts != nil {
		params := u.Query()
		if opts.Limit > 0 {
			params.Set("maxRecords", strconv.Itoa(opts.Limit))
		}
		if len(opts.Sort) > 0 {
			sortMaps := make([]map[string]string, len(opts.Sort))
			for i := range opts.Sort {
				if err := ms.Decode(&opts.Sort[i], &sortMaps[i]); err != nil {
					panic(err)
				}
			}

			for i, sortMap := range sortMaps {
				for sortField, sortVal := range sortMap {
					key := fmt.Sprintf("sort[%d][%s]", i, sortField)
					params.Set(key, sortVal)
				}
			}
		}
		u.RawQuery = params.Encode()
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.APIKey))
	res, err := c.HC.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Check response status code.
	if res.StatusCode != 200 {
		return fmt.Errorf("airtable: bad status code (%d)", res.StatusCode)
	}

	// Unmarshal response into v.
	var data struct {
		Records []struct {
			Fields interface{} `json:"fields"`
		} `json:"records"`
	}
	dec := json.NewDecoder(res.Body)
	if err = dec.Decode(&data); err != nil {
		return ess.AddCtx("airtable: decoding response body", err)
	}
	if err = res.Body.Close(); err != nil {
		return ess.AddCtx("airtable: closing response body", err)
	}

	records := make([]interface{}, len(data.Records))
	for i := range data.Records {
		records[i] = data.Records[i].Fields
	}

	mdec, err := MapDecoder(v)
	if err != nil {
		return ess.AddCtx("airtable: creating mapstructure decoder", err)
	}
	err = mdec.Decode(records)
	return ess.AddCtx("airtable: decoding records into receiver", err)
}

// MapDecoder produces a mapstructure.Decoder with DecodeHooks for parsing
// Airtable field values.
func MapDecoder(result interface{}) (*ms.Decoder, error) {
	config := ms.DecoderConfig{
		Result:     result,
		DecodeHook: ms.StringToTimeHookFunc(time.RFC3339),
	}
	return ms.NewDecoder(&config)
}
