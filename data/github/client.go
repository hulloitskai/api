package github

import (
	"context"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	errors "golang.org/x/xerrors"

	"github.com/google/go-github/v25/github"
	"github.com/stevenxie/api/pkg/api"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "github"

// A Client can access the GitHub API.
type Client struct {
	ghc   *github.Client
	httpc *http.Client

	currentUserLogin string
}

var (
	_ GistRepo              = (*Client)(nil)
	_ api.GitCommitsService = (*Client)(nil)
)

// New creates a new GitHub client.
//
// It reads GITHUB_TOKEN from the environment; if no such variable is found, an
// error will be returned.
func New() (*Client, error) {
	token := os.Getenv(strings.ToUpper(Namespace) + "_TOKEN")
	if token == "" {
		return nil, ErrBadEnvToken
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

// ErrBadEnvToken means that no 'GITHUB_TOKEN' environment variable was found.
var ErrBadEnvToken = errors.Errorf("github: no such environment variable "+
	"'%s_TOKEN'", strings.ToUpper(Namespace))
