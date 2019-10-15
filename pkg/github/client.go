package github

import (
	"context"
	"net/http"
	"os"

	"go.stevenxie.me/api/pkg/name"
	"golang.org/x/oauth2"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v25/github"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "github"

// A Client can access the GitHub API.
type Client struct {
	ghc              *github.Client
	httpc            *http.Client
	currentUserLogin string
}

// New creates a new GitHub client.
//
// It reads GITHUB_TOKEN from the environment; if no such variable is found, an
// error will be returned.
func New() (*Client, error) {
	var token string
	{
		var (
			key = name.EnvKey(Namespace, "TOKEN")
			ok  bool
		)
		if token, ok = os.LookupEnv(key); !ok {
			return nil, errors.Newf("github: no such environment variable '%s'", key)
		}
	}

	// Create authenticated http.Client.
	var (
		source = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), source)
	)
	return &Client{
		ghc:   github.NewClient(client),
		httpc: client,
	}, nil
}

// BaseURL returns GitHub's API base URL.
func (c *Client) BaseURL() string {
	url := c.ghc.BaseURL.String()
	return url[:len(url)-1]
}

// HTTP returns an authenticated http.Client that is authorized to make
// requests to the GitHub API.
func (c *Client) HTTP() *http.Client { return c.httpc }

// GitHub returns an authenticated github.Client.
func (c *Client) GitHub() *github.Client { return c.ghc }
