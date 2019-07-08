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
	// A SegmentsStreamer implements a geo.StreamingSegmentSource that
	// preloads recent locations data.
	SegmentsStreamer struct {
		streamer *PollStreamer
		log      *logrus.Logger

		stream chan struct {
			Segment *geo.Segment
			Err     error
		}

		mux      sync.Mutex
		segments []*geo.Segment
		err      error
	}

	// An SSOption configures a SegmentsStreamer.
	SSOption func(*SegmentsStreamer)
)

var _ geo.StreamingSegmentSource = (*SegmentsStreamer)(nil)

// NewSegmentsStreamer creates a new SegmentsStreamer.
func NewSegmentsStreamer(
	source geo.SegmentSource,
	interval time.Duration,
	opts ...SSOption,
) *SegmentsStreamer {
	var (
		action = func() (zero.Interface, error) { return source.RecentSegments() }
		ss     = &SegmentsStreamer{
			streamer: NewPollStreamer(action, interval),
			stream: make(chan struct {
				Segment *geo.Segment
				Err     error
			}),
			log: zero.Logger(),
		}
	)
	for _, opt := range opts {
		opt(ss)
	}
	go ss.startStreaming()
	return ss
}

// WithLSLogger configures a LocationPreloader's logger.
func WithLSLogger(log *logrus.Logger) SSOption {
	return func(rlp *SegmentsStreamer) { rlp.log = log }
}

func (ss *SegmentsStreamer) startStreaming() {
	for result := range ss.streamer.Stream() {
		var (
			segments []*geo.Segment
			err      error
		)

		switch v := result.(type) {
		case error:
			err = v
			ss.log.WithError(err).Error("Failed to load last seen position.")
		case []*geo.Segment:
			segments = v
		default:
			ss.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Errorf("stream: unexpected value '%s' from upstream")
		}

		// Write values to stream.
		ss.mux.Lock()
		if err != nil {
			ss.stream <- struct {
				Segment *geo.Segment
				Err     error
			}{Err: err}
		} else {
			for i, segment := range segments {
				if i >= len(ss.segments) {
					goto Send
				}
				if len(segment.Coordinates) == len(ss.segments[i].Coordinates) {
					continue
				}
			Send:
				ss.stream <- struct {
					Segment *geo.Segment
					Err     error
				}{Segment: segment}
			}
		}

		// Cache values.
		ss.segments = segments
		ss.err = err
		ss.mux.Unlock()
	}
}

// Stop stops the RecentLocationPreloader.
func (ss *SegmentsStreamer) Stop() { ss.streamer.Stop() }

// SegmentsStream returns a stream of location history segments.
func (ss *SegmentsStreamer) SegmentsStream() <-chan struct {
	Segment *geo.Segment
	Err     error
} {
	return ss.stream
}

// RecentSegments returns the authenticated user's recent location history.
func (ss *SegmentsStreamer) RecentSegments() ([]*geo.Segment, error) {
	ss.mux.Lock()
	defer ss.mux.Unlock()
	return ss.segments, nil
}
