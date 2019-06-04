package handler

import (
	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// NowPlayingHandler handles requests for the currently playing track on my
// Spotify account.
func NowPlayingHandler(
	svc api.NowPlayingService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		cplaying, err := svc.NowPlaying()
		if err != nil {
			log.WithError(err).Error("Failed to get currently playing track.")
			return errors.Errorf("getting currently playing track: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, cplaying)
	}
}
