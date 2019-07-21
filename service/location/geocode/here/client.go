package here

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "here"

type (
	// A Client can make requests to the Here API.
	Client struct {
		httpc    *http.Client
		id, code string
	}

	// An ClientConfig configures a Client.
	ClientConfig struct {
		HTTPClient *http.Client
	}
)

// NewClient creates a new Client. It reads the app code from the environment
// variable 'HERE_APP_CODE'.
func NewClient(appID string, opts ...func(*ClientConfig)) (*Client, error) {
	cfg := ClientConfig{
		HTTPClient: new(http.Client),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		envvar   = fmt.Sprintf("%s_APP_CODE", strings.ToUpper(Namespace))
		code, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("here: no such envvar '%s'", envvar)
	}

	return &Client{
		httpc: cfg.HTTPClient,
		id:    appID,
		code:  code,
	}, nil
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	params := req.URL.Query()
	params.Set("app_id", c.id)
	params.Set("app_code", c.code)
	req.URL.RawQuery = params.Encode()
	return c.httpc.Do(req)
}
