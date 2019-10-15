package spotify

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/zmb3/spotify"
	"golang.org/x/oauth2"

	"go.stevenxie.me/api/pkg/name"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "spotify"

// New creates a new Spotify client.
//
// It uses the environment variable 'SPOTIFY_TOKEN' as the refresh token; if no
// such variable is found, an error will be returned.
func New() (*spotify.Client, error) {
	var refresh string
	{
		var (
			key = name.EnvKey(Namespace, "TOKEN")
			ok  bool
		)
		if refresh, ok = os.LookupEnv(key); !ok {
			return nil, errors.Newf("spotify: no such environment variable '%s'", key)
		}
	}
	var (
		token  = &oauth2.Token{RefreshToken: refresh, TokenType: "Bearer"}
		client = spotify.NewAuthenticator("", "").NewClient(token)
	)
	return &client, nil
}
