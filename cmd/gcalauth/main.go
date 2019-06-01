package main

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"

	"github.com/stevenxie/api/internal/cmdutil"
	ess "github.com/unixpickle/essentials"
)

var (
	redirURL = "http://localhost:3000/callback"
	scopes   = []string{
		calendar.CalendarEventsReadonlyScope,
		calendar.CalendarReadonlyScope,
		calendar.CalendarSettingsReadonlyScope,
	}
)

func main() {
	cmdutil.PrepareEnv()
	var (
		config = oauth2.Config{
			RedirectURL:  redirURL,
			ClientID:     os.Getenv("GOOGLE_ID"),
			ClientSecret: os.Getenv("GOOGLE_SECRET"),
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		}
		randn = rand.NewSource(time.Now().UnixNano()).Int63()
		state = strconv.FormatInt(randn, 10)
	)

	// Declare redirect handler.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(
			w,
			r,
			config.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce),
			http.StatusTemporaryRedirect)
	})

	// Declare exchange handler.
	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.FormValue("state") != state {
			http.Error(w, "bad state", http.StatusBadRequest)
			return // break early
		}

		// Exchange code for token.
		token, err := config.Exchange(context.Background(), r.FormValue("code"))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to exchange code: %v", err),
				http.StatusInternalServerError)
			return // break early
		}

		// Check refresh token.
		if token.RefreshToken == "" {
			http.Error(w, "bad refresh token", http.StatusBadRequest)
			return // break early
		}

		io.WriteString(w, token.RefreshToken)
	})

	// Start server.
	if err := http.ListenAndServe(":3000", nil); err != nil {
		ess.Die("Error running server:", err)
	}
}
