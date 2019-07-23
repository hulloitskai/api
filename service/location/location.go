package location

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/pkg/stream"
	"go.stevenxie.me/api/pkg/zero"
)

type (
	// A HistoryPreloader implements a geo.SegmentSource that preloads recent
	// location history data.
	HistoryPreloader struct {
		streamer stream.Streamer
		log      logrus.FieldLogger

		mux      sync.Mutex
		segments HistorySegments
		err      error
	}

	// An HistoryPreloaderConfig configures a HistoryPreloader.
	HistoryPreloaderConfig struct {
		Logger logrus.FieldLogger
	}
)

var _ HistoryService = (*HistoryPreloader)(nil)

// NewHistoryPreloader creates a new HistoryPreloader.
func NewHistoryPreloader(
	svc HistoryService,
	interval time.Duration,
	opts ...func(*HistoryPreloaderConfig),
) *HistoryPreloader {
	cfg := HistoryPreloaderConfig{
		Logger: zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	sp := &HistoryPreloader{
		streamer: stream.NewPoller(
			func() (zero.Interface, error) {
				return svc.RecentHistory()
			},
			interval,
		),
		log: cfg.Logger,
	}
	go sp.populateCache()
	return sp
}

func (sp *HistoryPreloader) populateCache() {
	for result := range sp.streamer.Stream() {
		var (
			segments HistorySegments
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			sp.log.WithError(err).Error("Failed to load last seen position.")
		case HistorySegments:
			segments = v
		default:
			sp.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("location: unexpected upstream value '%s'", v)
		}

		sp.mux.Lock()
		sp.segments = segments
		sp.err = err
		sp.mux.Unlock()
	}
}

// Stop stops the HistoryPreloader.
func (sp *HistoryPreloader) Stop() { sp.streamer.Stop() }

// RecentHistory returns my recent location history.
func (sp *HistoryPreloader) RecentHistory() (HistorySegments, error) {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	return sp.segments, nil
}
