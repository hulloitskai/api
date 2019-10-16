package rescuetime

import (
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/pkg/httputil"
	"go.stevenxie.me/gopkg/name"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "rescuetime"

// NewClient creates a new Client.
//
// It reads 'RESCUETIME_KEY' (an API key) from the environment; if no such
// variable is found, an error will be returned.
func NewClient(opts ...httputil.BasicClientOption) (Client, error) {
	cfg := httputil.BasicClientConfig{
		HTTPClient: new(http.Client),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var key string
	{
		var (
			envKey = name.EnvKey(Namespace, "KEY")
			ok     bool
		)
		if key, ok = os.LookupEnv(envKey); !ok {
			return nil, errors.Newf(
				"rescuetime: no such environment variable '%s'",
				envKey,
			)
		}
	}

	return client{
		httpc: cfg.HTTPClient,
		key:   key,
	}, nil
}

type (
	// A Client can make authenticated requests to the RescueTime API.
	Client httputil.BasicClient

	client struct {
		httpc *http.Client
		key   string
	}
)

func (c client) Do(req *http.Request) (*http.Response, error) {
	// Set 'key' query parameter.
	params := req.URL.Query()
	params.Set("key", c.key)
	req.URL.RawQuery = params.Encode()

	// Perform request.
	return c.httpc.Do(req)
}
