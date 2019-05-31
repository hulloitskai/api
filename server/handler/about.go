package handler

import (
	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/about"
)

// AboutHandler responds with personal data.
func AboutHandler(l zerolog.Logger, svc about.InfoService) echo.HandlerFunc {
	return func(c echo.Context) error {
		info, err := svc.Info()
		if err != nil {
			l.Err(err).Msg("Failed to load info from store.")
			return errors.Errorf("loading info from store: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, info)
	}
}
