// Command spotifyauth initiates an OAuth2 flow for retrieving a Spotify
// refresh token for a particular scope.
package main

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/stevenxie/api/internal/cmdutil"
	ess "github.com/unixpickle/essentials"
	"github.com/zmb3/spotify"
)

const (
	redirURL = "http://localhost:3000/callback"
	scope    = spotify.ScopeUserReadCurrentlyPlaying
)

func main() {
	cmdutil.PrepareEnv()
	var (
		auth  = spotify.NewAuthenticator(redirURL, scope)
		randn = rand.NewSource(time.Now().UnixNano()).Int63()
		state = strconv.FormatInt(randn, 10)
	)

	// Declare redirect handler.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, auth.AuthURL(state),
			http.StatusTemporaryRedirect)
	})

	// Declare exchange handler.
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.Token(state, r)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to extract token: %v", err),
				http.StatusBadRequest)
			return // break early
		}
		io.WriteString(w, token.RefreshToken)
	})

	// Start server.
	if err := http.ListenAndServe(":3000", nil); err != nil {
		ess.Die("Error running server:", err)
	}
}
