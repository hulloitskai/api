package handler

import (
	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/music"
)

// NowPlayingHandler handles requests for the currently playing track on my
// Spotify account.
func NowPlayingHandler(
	l zerolog.Logger,
	svc music.CurrentlyPlayingService,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := svc.CurrentlyPlaying()
		if err != nil {
			l.Err(err).Msg("Failed to get currently playing track.")
			return errors.Errorf("getting currently playing track: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}
