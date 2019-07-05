package stream

import (
	"sync"
	"time"

	errors "golang.org/x/xerrors"

	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// NowPlayingStreamer implements a streaming interface over an
	// api.NowPlayingService.
	NowPlayingStreamer struct {
		streamer *PollStreamer
		log      *logrus.Logger

		stream chan api.MaybeNowPlaying

		mux    sync.Mutex
		latest api.MaybeNowPlaying
	}

	// An NPSOption configures a NowPlayingStreamer.
	NPSOption func(*NowPlayingStreamer)
)

// Ensure that a NowPlayingStreamer implements both api.NowPlayingService and
// api.NowPlayingStreamingService.
var (
	_ api.NowPlayingService          = (*NowPlayingStreamer)(nil)
	_ api.NowPlayingStreamingService = (*NowPlayingStreamer)(nil)
)

// WithNPSLogger adds an logger to a NowPlayingStreamer.
func WithNPSLogger(log *logrus.Logger) NPSOption {
	return func(nps *NowPlayingStreamer) { nps.log = log }
}

// NewNowPlayingStreamer creates a new NowPlayingStreamer.
func NewNowPlayingStreamer(
	svc api.NowPlayingService,
	interval time.Duration,
	opts ...NPSOption,
) *NowPlayingStreamer {
	var (
		action   = func() (zero.Interface, error) { return svc.NowPlaying() }
		streamer = &NowPlayingStreamer{
			streamer: NewPollStreamer(action, interval),
			stream:   make(chan api.MaybeNowPlaying),
			log:      zero.Logger(),
		}
	)
	for _, opt := range opts {
		opt(streamer)
	}
	go streamer.startStreaming()
	return streamer
}

func (nps *NowPlayingStreamer) startStreaming() {
	for result := range nps.streamer.Stream() {
		var maybe api.MaybeNowPlaying
		switch v := result.(type) {
		case error:
			maybe = api.MaybeNowPlaying{Err: v}
			nps.log.WithError(maybe.Err).Error("Failed to load now-playing data.")
		case *api.NowPlaying:
			maybe = api.MaybeNowPlaying{NowPlaying: v}
		default:
			nps.log.WithField("value", v).Error("Unexpected value from upstream.")
			maybe = api.MaybeNowPlaying{
				Err: errors.Errorf("stream: unexpected upstream value (%v)", v),
			}
		}

		// Safely write maybe to latest.
		nps.mux.Lock()
		nps.latest = maybe
		nps.mux.Unlock()

		// Write maybe to stream.
		nps.stream <- maybe
	}
}

// NowPlayingStream exposes a stream of NowPlaying objects.
func (nps *NowPlayingStreamer) NowPlayingStream() <-chan api.MaybeNowPlaying {
	return nps.stream
}

// NowPlaying returns the latest NowPlaying stream result.
func (nps *NowPlayingStreamer) NowPlaying() (*api.NowPlaying, error) {
	nps.mux.Lock()
	defer nps.mux.Unlock()
	return nps.latest.NowPlaying, nps.latest.Err
}

// Stop stops the NowPlayingStreamer.
func (nps *NowPlayingStreamer) Stop() { nps.streamer.Stop() }
