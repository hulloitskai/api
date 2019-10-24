package schedsvc

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/basic"
	"go.stevenxie.me/api/scheduling"
)

// NewService creates a new Service.
func NewService(
	cal scheduling.Calendar,
	zones location.TimeZoneService,
	opts ...basic.Option,
) scheduling.Service {
	cfg := basic.BuildConfig(opts...)
	return service{
		cal:   cal,
		zones: zones,
		log:   logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	cal   scheduling.Calendar
	zones location.TimeZoneService
	log   *logrus.Entry
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
	periods, err := svc.cal.BusyTimes(ctx, date)
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

func (svc service) BusyTimesToday(ctx context.Context) ([]scheduling.TimeSpan, error) {
	log := logutil.
		WithMethod(svc.log, service.BusyTimesToday).
		WithContext(ctx)

	log.Trace("Getting current time zone...")
	tz, err := svc.zones.CurrentTimeZone(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current time zone.")
		return nil, errors.Wrap(err, "schedsvc: get current time zone")
	}
	return svc.BusyTimes(ctx, time.Now().In(tz))
}
