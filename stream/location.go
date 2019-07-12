package stream

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A SegmentsPreloader implements a geo.SegmentSource that preloads recent
	// location history data.
	SegmentsPreloader struct {
		streamer *PollStreamer
		log      *logrus.Logger

		mux      sync.Mutex
		segments []*geo.Segment
		err      error
	}

	// An SPConfig configures a SegmentsPreloader.
	SPConfig struct{ Logger *logrus.Logger }
)

var _ geo.SegmentSource = (*SegmentsPreloader)(nil)

// NewSegmentsPreloader creates a new SegmentsPreloader.
func NewSegmentsPreloader(
	source geo.SegmentSource,
	interval time.Duration,
	opts ...func(*SPConfig),
) *SegmentsPreloader {
	cfg := SPConfig{Logger: zero.Logger()}
	for _, opt := range opts {
		opt(&cfg)
	}

	sp := &SegmentsPreloader{
		streamer: NewPollStreamer(
			func() (zero.Interface, error) { return source.RecentSegments() },
			interval,
		),
		log: cfg.Logger,
	}
	go sp.populateCache()
	return sp
}

func (sp *SegmentsPreloader) populateCache() {
	for result := range sp.streamer.Stream() {
		var (
			segments []*geo.Segment
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			sp.log.WithError(err).Error("Failed to load last seen position.")
		case []*geo.Segment:
			segments = v
		default:
			sp.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("stream: unexpected upstream value '%s'", v)
		}

		sp.mux.Lock()
		sp.segments = segments
		sp.err = err
		sp.mux.Unlock()
	}
}

// Stop stops the RecentLocationPreloader.
func (sp *SegmentsPreloader) Stop() { sp.streamer.Stop() }

// RecentSegments returns the authenticated user's recent location history.
func (sp *SegmentsPreloader) RecentSegments() ([]*geo.Segment, error) {
	sp.mux.Lock()
	defer sp.mux.Unlock()
	return sp.segments, nil
}
