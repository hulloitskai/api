package handler

import (
	"net/http"
	"time"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/httputil"
)

// AvailabilityHandler handles requests for availability information.
func AvailabilityHandler(
	svc api.AvailabilityService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Derive timezone.
		var (
			timezone *time.Location
			err      error
		)
		if timezonep := c.QueryParam("timezone"); timezonep != "" {
			if timezone, err = time.LoadLocation(timezonep); err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Errorf("bad paremter 'timezone': %w", err)
			}
		} else {
			if timezone, err = svc.Timezone(); err != nil {
				log.WithError(err).Error("Failed to load default timezone.")
				return errors.Errorf("failed to load default timezone: %w", err)
			}
		}

		// Derive date.
		date := time.Now().In(timezone)
		if datep := c.QueryParam("date"); datep != "" {
			date, err = time.ParseInLocation("2006-01-02", datep, timezone)
			if err != nil {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.Errorf("bad parameter 'date': %w", err)
			}
		}

		// Check busy periods.
		busyPeriods, err := svc.BusyPeriods(date)
		if err != nil {
			log.WithError(err).Error("Failed to load availability.")
			return err
		}

		return jsonPretty(c, struct {
			Busy     []*api.TimePeriod `json:"busy"`
			Timezone string            `json:"timezone"`
		}{
			Busy:     busyPeriods,
			Timezone: timezone.String(),
		})
	}
}
