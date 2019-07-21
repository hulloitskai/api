package github

import (
	"context"
	"net/http"
	"os"
	"strings"

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
	token := os.Getenv(strings.ToUpper(Namespace) + "_TOKEN")
	if token == "" {
		return nil, ErrNoToken
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

// ErrNoToken means that no 'GITHUB_TOKEN' environment variable was found.
var ErrNoToken = errors.Newf(
	"github: no such environment variable '%s_TOKEN'",
	strings.ToUpper(Namespace),
)

// BaseURL returns GitHub's API base URL.
func (c *Client) BaseURL() string {
	url := c.ghc.BaseURL.String()
	return url[:len(url)-1]
}

// HTTPClient returns an authenticated http.Client that is authorized to make
// requests to the GitHub API.
func (c *Client) HTTPClient() *http.Client { return c.httpc }

// GHClient returns an authenticated github.Client.
func (c *Client) GHClient() *github.Client { return c.ghc }
