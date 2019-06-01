package handler

import (
	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/api"
)

// AboutHandler responds with personal data.
func AboutHandler(svc api.AboutService, l zerolog.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		about, err := svc.About()
		if err != nil {
			l.Err(err).Msg("Failed to load about info.")
			return errors.Errorf("fetching about info: %w", err)
		}

		// Send info as JSON.
		return jsonPretty(c, about)
	}
}
