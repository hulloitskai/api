package here

import (
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/v2/pkg/httputil"
	"go.stevenxie.me/gopkg/name"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "here"

type (
	// A Client can make requests to the Here API.
	Client httputil.BasicClient

	client struct {
		httpc    *http.Client
		id, code string
	}
)

// NewClient creates a new Client. It reads the app code from the environment
// variable 'HERE_APP_CODE'.
func NewClient(
	appID string,
	opts ...httputil.BasicClientOption,
) (Client, error) {
	cfg := httputil.BasicClientConfig{
		HTTPClient: new(http.Client),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var code string
	{
		var (
			key = name.EnvKey(Namespace, "APP_CODE")
			ok  bool
		)
		if code, ok = os.LookupEnv(key); !ok {
			return nil, errors.Newf("here: no such environment variable '%s'", key)
		}
	}

	return client{
		httpc: cfg.HTTPClient,
		id:    appID,
		code:  code,
	}, nil
}

// Do sends an HTTP request and returns an HTTP response, following policy
// (such as redirects, cookies, auth) as configured on the client.
func (c client) Do(req *http.Request) (*http.Response, error) {
	// Set query params.
	ps := req.URL.Query()
	ps.Set("app_id", c.id)
	ps.Set("app_code", c.code)
	req.URL.RawQuery = ps.Encode()

	// Perform request.
	return c.httpc.Do(req)
}
