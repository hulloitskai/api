package airtable

import (
	"fmt"
	"net/http"
	"os"

	"go.stevenxie.me/api/pkg/httputil"
	"go.stevenxie.me/gopkg/name"

	"github.com/cockroachdb/errors"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "airtable"

type (
	// A Client can make authenticated requests to the Airtable API.
	Client httputil.BasicClient

	client struct {
		httpc *http.Client
		token string
	}
)

// NewClient creates a new Client.
func NewClient(opts ...httputil.BasicClientOption) (Client, error) {
	cfg := httputil.BasicClientConfig{
		HTTPClient: new(http.Client),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var (
		envvar    = name.EnvKey(Namespace, "API_KEY")
		token, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("airtable: no such envvar '%s'", envvar)
	}

	c := client{
		httpc: cfg.HTTPClient,
		token: token,
	}
	return c, nil
}

// Do sends an authenticated HTTP request.
func (c client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return c.httpc.Do(req)
}
