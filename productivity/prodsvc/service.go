package prodsvc

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/pkg/basic"
	"go.stevenxie.me/api/v2/productivity"
)

// NewService creates a new Services from a RecordService.
func NewService(
	records productivity.RecordSource,
	zones location.TimeZoneService,
	opts ...basic.Option,
) productivity.Service {
	cfg := basic.BuildOptions(opts...)
	return service{
		records: records,
		zones:   zones,
		log:     logutil.WithComponent(cfg.Logger, (*service)(nil)),
		tracer:  cfg.Tracer,
	}
}

type service struct {
	records productivity.RecordSource
	zones   location.TimeZoneService

	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ productivity.Service = (*service)(nil)

func (svc service) GetProductivity(
	ctx context.Context,
	date time.Time,
) (*productivity.Productivity, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.GetProductivity),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.GetProductivity),
		"date":            date,
	}).WithContext(ctx)

	// Get records.
	log.Trace("Getting records...")
	recs, err := svc.records.GetRecords(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to get records.")
		return nil, err
	}
	log.WithField("records", recs).Trace("Got records.")

	// Return early if no records.
	if len(recs) == 0 {
		return &productivity.Productivity{
			Records: []productivity.Record{},
		}, nil
	}

	// Sort records by weight.
	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Category.Weight() < recs[j].Category.Weight()
	})
	log.
		WithField("records", recs).
		Trace("Sorted records by weight.")

	// Compute score.
	var score uint
	{
		var totalSecs uint
		for _, r := range recs {
			secs := uint(math.Round(r.Duration.Seconds()))
			score += r.Category.Weight() * secs
			totalSecs += secs
		}
		score = uint(math.Round(float64(score) / float64(totalSecs*4) * 100))
	}
	log.
		WithField("score", score).
		Trace("Computed productivity score.")

	return &productivity.Productivity{
		Records: recs,
		Score:   &score,
	}, nil
}

func (svc service) CurrentProductivity(ctx context.Context) (
	*productivity.Productivity, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.CurrentProductivity),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.CurrentProductivity).
		WithContext(ctx)

	log.Trace("Getting current time zone.")
	zone, err := svc.zones.CurrentTimeZone(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current time zone.")
		return nil, errors.Wrap(err, "prodsvc: getting current time zone")
	}
	return svc.GetProductivity(ctx, time.Now().In(zone))
}
