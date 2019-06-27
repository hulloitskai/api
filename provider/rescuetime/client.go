package rescuetime

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/stevenxie/api/pkg/api"
	errors "golang.org/x/xerrors"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "rescuetime"

type (
	// Client can access the RescueTime API.
	Client struct {
		httpc    *http.Client
		key      string
		timezone *time.Location
	}

	// An Option configures a Client.
	Option func(c *Client)
)

var _ api.ProductivityService = (*Client)(nil)

// New creates a new Client.
//
// It reads RESCUETIME_KEY (an API key) from the environment; if no such
// variable is found, an error will be returned.
func New(opts ...Option) (*Client, error) {
	key := os.Getenv(strings.ToUpper(Namespace) + "_KEY")
	if key == "" {
		return nil, ErrBadEnvKey
	}

	c := &Client{
		httpc: new(http.Client),
		key:   key,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// WithTimezone configures the timezone that the Client will use to make
// time/date queries.
func WithTimezone(tz *time.Location) Option {
	return func(c *Client) { c.timezone = tz }
}

// ErrBadEnvKey means that no 'RESCUETIME_KEY' environment variable was found.
var ErrBadEnvKey = errors.Errorf("rescuetime: no such environment variable "+
	"'%s_KEY'", strings.ToUpper(Namespace))
