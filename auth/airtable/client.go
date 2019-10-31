package airtable

import (
	"fmt"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/v2/pkg/httputil"
	"go.stevenxie.me/gopkg/name"
)

// Namespace is the package namespace used for things like envvar prefixes.
const Namespace = "airtable"

// NewClient creates a new Client.
func NewClient(opts ...httputil.BasicClientOption) (Client, error) {
	opt := httputil.BasicClientOptions{
		HTTPClient: new(http.Client),
	}
	for _, apply := range opts {
		apply(&opt)
	}

	var (
		envvar    = name.EnvKey(Namespace, "API_KEY")
		token, ok = os.LookupEnv(envvar)
	)
	if !ok {
		return nil, errors.Newf("airtable: no such envvar '%s'", envvar)
	}

	c := client{
		httpc: opt.HTTPClient,
		token: token,
	}
	return c, nil
}

// Do sends an authenticated HTTP request.
func (c client) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	return c.httpc.Do(req)
}

type (
	// A Client can make authenticated requests to the Airtable API.
	Client httputil.BasicClient

	client struct {
		httpc *http.Client
		token string
	}
)
