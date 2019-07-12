package airtable

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
)

const (
	// Namespace is the package namespace used for things like envvar prefixes.
	Namespace = "airtable"

	baseURL = "https://api.airtable.com/v0"
)

type (
	// A Client can interact with the Airtable API.
	Client struct {
		httpc *http.Client
		token string
	}

	// A ClientOption configures a Client.
	ClientOption func(*Client)
)

// NewClient creates a new Client.
func NewClient(opts ...ClientOption) (*Client, error) {
	var (
		envvar    = fmt.Sprintf("%s_API_KEY", strings.ToUpper(Namespace))
		token, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("airtable: no such envvar '%s'", envvar)
	}

	c := &Client{
		httpc: new(http.Client),
		token: token,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// WithHTTPClient configures a Client to make network requests using httpc.
func WithHTTPClient(httpc *http.Client) ClientOption {
	return func(c *Client) { c.httpc = httpc }
}

// Do sends an authenticated HTTP request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return c.httpc.Do(req)
}

// Get sends an authenticated HTTP GET request.
func (c *Client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}
