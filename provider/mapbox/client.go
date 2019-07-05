package mapbox

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/stevenxie/api/pkg/geo"
	errors "golang.org/x/xerrors"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "mapbox"

const baseURL = "https://api.mapbox.com"

type (
	// A Client can interact with the MapBox API.
	Client struct {
		httpc *http.Client
		token string
	}

	// An Option configures a Client.
	Option func(*Client)
)

var _ geo.Geocoder = (*Client)(nil)

// New creates a new Client, with a token read from the environment (as
// 'MAPBOX_TOKEN').
func New(opts ...Option) (*Client, error) {
	var (
		envvar    = fmt.Sprintf("%s_TOKEN", strings.ToUpper(Namespace))
		token, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Errorf("mapbox: no such envvar '%s'", envvar)
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

// WithHTTPClient configures a Client to make HTTP requests with httpc.
func WithHTTPClient(httpc *http.Client) Option {
	return func(c *Client) { c.httpc = httpc }
}

func (c *Client) beginQuery(url *url.URL) url.Values {
	params := url.Query()
	params.Set("access_token", c.token)
	return params
}
