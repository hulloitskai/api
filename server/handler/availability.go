package handler

import (
	"net/http"
	"time"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// AvailabilityHandler handles requests for availability information.
func AvailabilityHandler(
	svc api.AvailabilityService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		loc, err := svc.Timezone()
		if err != nil {
			log.WithError(err).Error("Failed to determine timezone.")
			return errors.Errorf("failed to determine timezone: %w", err)
		}

		// Derive date.
		date := time.Now().In(loc)
		if datep := c.QueryParam("date"); datep != "" {
			date, err = time.ParseInLocation("2006-01-02", datep, loc)
			if err != nil {
				setRequestStatusCode(c, http.StatusBadRequest)
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
			Timezone: loc.String(),
		})
	}
}
