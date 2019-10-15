package locsvc

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/api/location/geocode/geoutil"
	"go.stevenxie.me/api/pkg/svcutil"
)

// NewHistoryService creates a HistoryService from a location.SegmentSource and
// a geocode.Geocoder.
func NewHistoryService(
	src location.SegmentSource,
	geo geocode.Geocoder,
	opts ...svcutil.BasicOption,
) location.HistoryService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &historyService{
		SegmentSource: src,
		geo:           geo,
		log:           cfg.Logger,
	}
}

type historyService struct {
	location.SegmentSource
	geo geocode.Geocoder
	log *logrus.Entry

	mux sync.Mutex
	loc *time.Location
}

var _ location.HistoryService = (*historyService)(nil)

func (svc *historyService) RecentHistory(ctx context.Context) (
	[]location.HistorySegment, error) {
	log := svc.log.
		WithField("method", "RecentHistory").
		WithContext(ctx)

	// Get current time, in my current time-zone (if applicable).
	now := time.Now()
	svc.mux.Lock()
	if svc.loc != nil {
		now = now.In(svc.loc)
	}
	svc.mux.Unlock()
	log = log.WithField("current_time", now)

	// Try loading today's segments.
	segs, err := svc.SegmentSource.GetHistory(ctx, now)
	if err != nil {
		log.WithError(err).Error("Failed to load today's history segments.")
		return nil, err
	}
	log = log.WithField("segments", segs)

	if len(segs) != 0 {
		log.Trace("Loaded today's history segments.")
	} else {
		// Try loading yesterday's segments.
		log.Trace("No history yet for today, loading yesterday's...")
		segs, err = svc.SegmentSource.GetHistory(ctx, now.Add(-24*time.Hour))
		if err != nil {
			log.
				WithError(err).
				Error("Failed to load yesterday's history segments.")
			return nil, err
		}
		log.Trace("Loaded yesterday's history segments.")
	}

	if len(segs) == 0 {
		segs = []location.HistorySegment{}
		log.Warn("No history segments were found.")
	} else {
		// Derive time location for future queries.
		go svc.deriveTimeLocation(ctx, &segs[0])
	}
	return segs, nil
}

func (svc *historyService) deriveTimeLocation(
	ctx context.Context,
	seg *location.HistorySegment,
) {
	log := svc.log.WithFields(logrus.Fields{
		"method":  "deriveTimeLocation",
		"segment": seg,
	}).WithContext(ctx)

	coords := latestCoordinates(seg)
	if coords == nil {
		log.Warn("History segment contains no coordinates.")
		return
	}
	log = log.WithField("coordinates", coords)

	loc, err := geoutil.TimeLocation(ctx, svc.geo, *coords)
	if err != nil {
		log.
			WithError(err).
			Error("Failed to determine time location for coordinates.")
		return
	}
	log = log.WithField("location", loc)

	svc.mux.Lock()
	svc.loc = loc
	svc.mux.Unlock()
	log.Trace("Updated time location for future requests.")
}
