package here

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/geo"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "here"

type (
	// A Client can make requests to the Here API.
	Client struct {
		httpc    *http.Client
		id, code string
	}

	// An Option configures a Client.
	Option func(*Client)
)

var _ geo.Geocoder = (*Client)(nil)

// New creates a new Client. It reads the app code from the environment
// variable 'HERE_APP_CODE'.
func New(appID string, opts ...Option) (*Client, error) {
	var (
		envvar   = fmt.Sprintf("%s_APP_CODE", strings.ToUpper(Namespace))
		code, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("here: no such envvar '%s'", envvar)
	}

	c := &Client{
		httpc: new(http.Client),
		id:    appID,
		code:  code,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// WithHTTPClient configures a Client to make HTTP requests with httpc.
func WithHTTPClient(httpc *http.Client) Option {
	return func(c *Client) { c.httpc = httpc }
}

func (c *Client) beginQuery(url *url.URL) url.Values {
	params := url.Query()
	params.Set("app_id", c.id)
	params.Set("app_code", c.code)
	return params
}
