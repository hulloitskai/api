package stream

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	errors "golang.org/x/xerrors"

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

	// An SSOption configures a SegmentsPreloader.
	SSOption func(*SegmentsPreloader)
)

var _ geo.SegmentSource = (*SegmentsPreloader)(nil)

// NewSegmentsPreloader creates a new SegmentsPreloader.
func NewSegmentsPreloader(
	source geo.SegmentSource,
	interval time.Duration,
	opts ...SSOption,
) *SegmentsPreloader {
	var (
		action = func() (zero.Interface, error) { return source.RecentSegments() }
		ss     = &SegmentsPreloader{
			streamer: NewPollStreamer(action, interval),
			log:      zero.Logger(),
		}
	)
	for _, opt := range opts {
		opt(ss)
	}
	go ss.populateCache()
	return ss
}

// WithSPLogger configures a LocationPreloader's logger.
func WithSPLogger(log *logrus.Logger) SSOption {
	return func(sp *SegmentsPreloader) { sp.log = log }
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
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
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
