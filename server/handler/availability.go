package handler

import (
	"net/http"
	"time"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/pkg/httputil"
	"go.stevenxie.me/api/service/availability"
)

// AvailabilityHandler handles requests for availability information.
func AvailabilityHandler(
	svc availability.Service,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Derive timezone.
		var (
			timezone *time.Location
			err      error
		)
		const timezoneParamName = "timezone"
		if timezonep := c.QueryParam(timezoneParamName); timezonep != "" {
			if timezone, err = time.LoadLocation(timezonep); err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Wrapf(err, "bad parameter '%s'", timezoneParamName)
			}
		} else {
			if timezone, err = svc.Timezone(); err != nil {
				log.WithError(err).Error("Failed to load default timezone.")
				return errors.Wrap(err, "failed to load default timezone")
			}
		}

		// Derive date.
		date := time.Now().In(timezone)
		const dateParamName = "date"
		if datep := c.QueryParam(dateParamName); datep != "" {
			date, err = time.ParseInLocation("2006-01-02", datep, timezone)
			if err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Wrapf(err, "bad parameter '%s'", dateParamName)
			}
		}

		// Check busy periods.
		busyPeriods, err := svc.BusyPeriods(date)
		if err != nil {
			log.WithError(err).Error("Failed to load availability.")
			return err
		}

		return jsonPretty(c, struct {
			Busy     availability.Periods `json:"busy"`
			Timezone string               `json:"timezone"`
		}{
			Busy:     busyPeriods,
			Timezone: timezone.String(),
		})
	}
}
