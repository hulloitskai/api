package rescuetime

import (
	"net/http"
	"os"
	"strings"

	"github.com/stevenxie/api/pkg/api"
	errors "golang.org/x/xerrors"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "rescuetime"

// Client can access the RescueTime API.
type Client struct {
	httpc *http.Client
	key   string
}

var _ api.ProductivityService = (*Client)(nil)

// New creates a new Client.
//
// It reads RESCUETIME_KEY (an API key) from the environment; if no such
// variable is found, an error will be returned.
func New() (*Client, error) {
	key := os.Getenv(strings.ToUpper(Namespace) + "_KEY")
	if key == "" {
		return nil, ErrBadEnvKey
	}

	return &Client{
		httpc: new(http.Client),
		key:   key,
	}, nil
}

// ErrBadEnvKey means that no 'RESCUETIME_KEY' environment variable was found.
var ErrBadEnvKey = errors.Errorf("rescuetime: no such environment variable "+
	"'%s_KEY'", strings.ToUpper(Namespace))
