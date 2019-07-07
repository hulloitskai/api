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
	// A RecentLocationsPreloader implements a geo.RecentLocationsService that
	// preloads recent locations data.
	RecentLocationsPreloader struct {
		streamer *PollStreamer
		log      *logrus.Logger

		mux      sync.Mutex
		segments []*geo.Segment
		err      error
	}

	// An RLPOption configures a RecentLocationsPreloader.
	RLPOption func(*RecentLocationsPreloader)
)

var _ geo.RecentLocationsService = (*RecentLocationsPreloader)(nil)

// NewRecentLocationsPreloader creates a new RecentLocationsPreloader.
func NewRecentLocationsPreloader(
	locations geo.RecentLocationsService,
	interval time.Duration,
	opts ...RLPOption,
) *RecentLocationsPreloader {
	rlp := &RecentLocationsPreloader{
		log: zero.Logger(),
	}
	for _, opt := range opts {
		opt(rlp)
	}

	// Configure streamer.
	action := func() (zero.Interface, error) { return locations.RecentSegments() }
	rlp.streamer = NewPollStreamer(action, interval)

	go rlp.populateCache()
	return rlp
}

// WithLSLogger configures a LocationPreloader's logger.
func WithLSLogger(log *logrus.Logger) RLPOption {
	return func(rlp *RecentLocationsPreloader) { rlp.log = log }
}

func (rlp *RecentLocationsPreloader) populateCache() {
	for result := range rlp.streamer.Stream() {
		var (
			segments []*geo.Segment
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			rlp.log.WithError(err).Error("Failed to load last seen position.")
		case []*geo.Segment:
			segments = v
		default:
			rlp.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
		}

		rlp.mux.Lock()
		rlp.segments = segments
		rlp.err = err
		rlp.mux.Unlock()
	}
}

// Stop stops the RecentLocationPreloader.
func (rlp *RecentLocationsPreloader) Stop() { rlp.streamer.Stop() }

// RecentSegments returns the authenticated user's recent location history.
func (rlp *RecentLocationsPreloader) RecentSegments() ([]*geo.Segment, error) {
	rlp.mux.Lock()
	defer rlp.mux.Unlock()
	return rlp.segments, nil
}
