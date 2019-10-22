package schedsvc

import (
	"context"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/pkg/basic"
	"go.stevenxie.me/api/scheduling"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewService creates a new Service.
func NewService(
	cal scheduling.Calendar,
	opts ...basic.Option,
) scheduling.Service {
	cfg := basic.BuildConfig(opts...)
	return service{
		cal: cal,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	cal scheduling.Calendar
	log *logrus.Entry
}

var _ scheduling.Service = (*service)(nil)

func (svc service) BusyTimes(
	ctx context.Context,
	date time.Time,
) ([]scheduling.TimeSpan, error) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.BusyTimes),
		"date":            date,
	}).WithContext(ctx)

	log.Trace("Getting busy times from calendar...")
	periods, err := svc.cal.RawBusyTimes(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to load busy times from calendar.")
		return nil, err
	}
	log.
		WithField("periods", periods).
		Trace("Loaded busy times from calendar.")

	// Sort periods.
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Before(&periods[j])
	})
	log.
		WithField("periods", periods).
		Trace("Sorted busy times by time.")
	return periods, nil
}
