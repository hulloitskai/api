package handler

import (
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/spotify"
	errors "golang.org/x/xerrors"
)

// NowPlayingHandler handles requests for the currently playing track on my
// Spotify account.
func NowPlayingHandler(
	l zerolog.Logger,
	svc spotify.CurrentlyPlayingService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := svc.GetCurrentlyPlaying()
		if err != nil {
			l.Err(err).Msg("Failed to get currently playing track.")
			return errors.Errorf("getting currently playing track: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}
