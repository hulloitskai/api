package prodsvc

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/api/productivity"
)

// NewService creates a new Services from a RecordService.
func NewService(
	records productivity.RecordSource,
	zones location.TimeZoneService,
	opts ...svcutil.BasicOption,
) productivity.Service {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		records: records,
		zones:   zones,
		log:     cfg.Logger,
	}
}

type service struct {
	records productivity.RecordSource
	zones   location.TimeZoneService
	log     *logrus.Entry
}

var _ productivity.Service = (*service)(nil)

func (svc service) GetProductivity(
	ctx context.Context,
	date time.Time,
) (*productivity.Productivity, error) {
	log := svc.log.WithFields(logrus.Fields{
		"method": "GetProductivity",
		"date":   date,
	}).WithContext(ctx)

	// Get records.
	recs, err := svc.records.GetRecords(ctx, date)
	if err != nil {
		log.WithError(err).Error("Failed to get records.")
		return nil, err
	}

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

	return &productivity.Productivity{
		Records: recs,
		Score:   &score,
	}, nil
}

func (svc service) CurrentProductivity(ctx context.Context) (
	*productivity.Productivity, error) {
	zone, err := svc.zones.CurrentTimeZone(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "prodsvc: getting current time zone")
	}
	return svc.GetProductivity(ctx, time.Now().In(zone))
}
