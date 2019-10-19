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
	src scheduling.BusySource,
	opts ...basic.Option,
) scheduling.Service {
	cfg := basic.BuildConfig(opts...)
	return service{
		src: src,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	src scheduling.BusySource
	log *logrus.Entry
}

var _ scheduling.Service = (*service)(nil)

func (svc service) BusyPeriods(
	ctx context.Context,
	date time.Time,
) ([]scheduling.TimePeriod, error) {
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.BusyPeriods),
		"date":            date,
	}).WithContext(ctx)

	log.Trace("Getting busy periods from source...")
	periods, err := svc.src.RawBusyPeriods(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to load busy periods from source.")
		return nil, err
	}
	log.
		WithField("periods", periods).
		Trace("Loaded busy periods from source.")

	// Sort periods.
	sort.Slice(periods, func(i, j int) bool {
		return periods[i].Before(&periods[j])
	})
	log.
		WithField("periods", periods).
		Trace("Sorted busy periods by time.")
	return periods, nil
}
