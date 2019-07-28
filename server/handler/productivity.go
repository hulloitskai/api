package handler

import (
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/service/productivity"
)

// ProductivityHandler handles requests for productivity metrics.
func ProductivityHandler(
	svc productivity.Service,
	log logrus.FieldLogger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		productivity, err := svc.CurrentProductivity()
		if err != nil {
			log.WithError(err).Error("Failed to load current productivity data.")
			return err
		}

		// Write productivity as JSON.
		return jsonPretty(c, productivity)
	}
}
