package spotify

import (
	"os"
	"strings"

	"golang.org/x/oauth2"
	errors "golang.org/x/xerrors"

	abstract "github.com/stevenxie/api/pkg/spotify"
	"github.com/zmb3/spotify"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "spotify"

// A Client can access the Spotify API.
type Client struct{ sc spotify.Client }

var _ abstract.CurrentlyPlayingService = (*Client)(nil)

// New creates a new Spotify client.
//
// It reads SPOTIFY_TOKEN (the refresh token) from the environment; if no such
// variable is found, an error will be returned.
func New() (*Client, error) {
	refresh := os.Getenv(strings.ToUpper(Namespace) + "_TOKEN")
	if refresh == "" {
		return nil, ErrBadEnvToken
	}

	var (
		token  = &oauth2.Token{RefreshToken: refresh, TokenType: "Bearer"}
		client = spotify.NewAuthenticator("", "").NewClient(token)
	)
	return &Client{sc: client}, nil
}

// ErrBadEnvToken means that no 'GITHUB_TOKEN' environment variable was found.
var ErrBadEnvToken = errors.New("spotify: no such environment variable " +
	"'SPOTIFY_TOKEN'")
