package airtable

import (
	"net/http"
	"net/http/cookiejar"

	errors "golang.org/x/xerrors"
)

// Client is capable of fetching data from the Airtable API.
type Client struct {
	httpc *http.Client

	apiKey, baseID string
}

// NewClient creates a new Airtable client.
func NewClient(apiKey, baseID string) (*Client, error) {
	return NewClientCustom(apiKey, baseID, nil)
}

// NewClientCustom creates a new Airtable client using a custom http.Client.
func NewClientCustom(apiKey, baseID string, client *http.Client) (*Client,
	error) {
	if apiKey == "" {
		return nil, errors.New("airtable: empty API key")
	}
	if baseID == "" {
		return nil, errors.New("airtable: empty base ID")
	}
	if client == nil {
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, errors.Errorf("airtable: creating cookiejar: %w", err)
		}
		client = &http.Client{Jar: jar}
	}
	return &Client{
		httpc:  client,
		apiKey: apiKey,
		baseID: baseID,
	}, nil
}
