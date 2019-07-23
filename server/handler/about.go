package handler

import (
	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/service/about"
)

// AboutHandler responds with personal data.
func AboutHandler(svc about.Service, log *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		info, err := svc.Info()
		if err != nil {
			log.WithError(err).Error("Failed to load info.")
			return errors.Wrap(err, "fetching info")
		}

		// Send info as JSON.
		return jsonPretty(c, info)
	}
}
