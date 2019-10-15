package httputil

import "net/http"

type (
	// A BasicClient can perform HTTP requests.
	BasicClient interface {
		Do(req *http.Request) (*http.Response, error)
	}

	// A Client is like a BasicClient, except it contains tasty shortcuts for
	// GET requests.
	Client interface {
		BasicClient
		Get(url string) (*http.Response, error)
	}
)

// ClientFromBasic creates a Client from a BasicClient.
func ClientFromBasic(bc BasicClient) Client {
	return client{bc}
}

type client struct {
	BasicClient
}

var _ Client = (*client)(nil)

func (c client) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return c.BasicClient.Do(req)
}

// WithBasicHTTPClient configures a BasicClient to make HTTP requests using c.
func WithBasicHTTPClient(c *http.Client) BasicClientOption {
	return func(cfg *BasicClientConfig) { cfg.HTTPClient = c }
}

type (
	// A BasicClientConfig configures a BasicClient.
	BasicClientConfig struct {
		HTTPClient *http.Client
	}

	// A BasicClientOption modifies a BasicClientConfig.
	BasicClientOption func(*BasicClientConfig)
)
