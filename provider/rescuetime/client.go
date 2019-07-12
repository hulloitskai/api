package rescuetime

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/api"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "rescuetime"

type (
	// Client can access the RescueTime API.
	Client struct {
		httpc    *http.Client
		timezone *time.Location
		key      string
	}

	// An ClientConfig configures a Client.
	ClientConfig struct {
		HTTPClient *http.Client
		Timezone   *time.Location
	}
)

var _ api.ProductivityService = (*Client)(nil)

// New creates a new Client.
//
// It reads RESCUETIME_KEY (an API key) from the environment; if no such
// variable is found, an error will be returned.
func New(opts ...func(*ClientConfig)) (*Client, error) {
	cfg := ClientConfig{HTTPClient: new(http.Client)}
	for _, opt := range opts {
		opt(&cfg)
	}

	key := os.Getenv(strings.ToUpper(Namespace) + "_KEY")
	if key == "" {
		return nil, ErrBadEnvKey
	}

	return &Client{
		httpc:    cfg.HTTPClient,
		timezone: cfg.Timezone,
		key:      key,
	}, nil
}

// ErrBadEnvKey means that no 'RESCUETIME_KEY' environment variable was found.
var ErrBadEnvKey = errors.Newf(
	"rescuetime: no such environment variable '%s_KEY'",
	strings.ToUpper(Namespace),
)
