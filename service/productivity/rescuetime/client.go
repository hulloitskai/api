package rescuetime

import (
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "rescuetime"

type (
	// A Client can access the RescueTime API.
	Client struct {
		httpc *http.Client
		key   string
	}

	// An ClientConfig configures a Client.
	ClientConfig struct {
		HTTPClient *http.Client
	}
)

// NewClient creates a new Client.
//
// It reads RESCUETIME_KEY (an API key) from the environment; if no such
// variable is found, an error will be returned.
func NewClient(opts ...func(*ClientConfig)) (*Client, error) {
	cfg := ClientConfig{
		HTTPClient: new(http.Client),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	key := os.Getenv(strings.ToUpper(Namespace) + "_KEY")
	if key == "" {
		return nil, errors.Newf(
			"rescuetime: no such environment variable '%s_KEY'",
			strings.ToUpper(Namespace),
		)
	}

	return &Client{
		httpc: cfg.HTTPClient,
		key:   key,
	}, nil
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Set 'key' query parameter.
	params := req.URL.Query()
	params.Set("key", c.key)
	req.URL.RawQuery = params.Encode()

	// Perform request.
	return c.httpc.Do(req)
}
