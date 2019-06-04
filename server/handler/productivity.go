package handler

import (
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// ProductivityHandler handles requests for productivity metrics.
func ProductivityHandler(
	svc api.ProductivityService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		prod, err := svc.CurrentProductivity()
		if err != nil {
			log.WithError(err).Error("Failed to load current productivity data.")
			return err
		}

		// Write productivity as JSON.
		return jsonPretty(c, prod)
	}
}
