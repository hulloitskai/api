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

	// A ClientConfig configures a Client.
	ClientConfig struct{ HTTPClient *http.Client }
)

// NewClient creates a new Client.
func NewClient(opts ...func(*ClientConfig)) (*Client, error) {
	cfg := ClientConfig{HTTPClient: new(http.Client)}
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		envvar    = fmt.Sprintf("%s_API_KEY", strings.ToUpper(Namespace))
		token, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("airtable: no such envvar '%s'", envvar)
	}

	c := &Client{
		httpc: cfg.HTTPClient,
		token: token,
	}
	return c, nil
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
