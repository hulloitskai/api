package handler

import (
	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// AboutHandler responds with personal data.
func AboutHandler(svc api.AboutService, log *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		about, err := svc.About()
		if err != nil {
			log.WithError(err).Error("Failed to load about info.")
			return errors.Wrap(err, "fetching about info")
		}

		// Send info as JSON.
		return jsonPretty(c, about)
	}
}
