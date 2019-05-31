package handler

import (
	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/metrics"
)

// ProductivityHandler handles requests for productivity metrics.
func ProductivityHandler(
	svc metrics.ProductivityService,
	l zerolog.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		prod, err := svc.CurrentProductivity()
		if err != nil {
			l.Err(err).Msg("Failed to load current productivity data.")
			return err
		}

		// Write productivity as JSON.
		return jsonPretty(c, prod)
	}
}
