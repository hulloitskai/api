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
	// MusicStreamer implements a streaming interface over an
	// api.MusicService.
	MusicStreamer struct {
		streamer *PollStreamer
		log      *logrus.Logger

		stream chan api.MaybeNowPlaying

		mux    sync.Mutex
		latest api.MaybeNowPlaying
	}

	// An NPSOption configures a NowPlayingStreamer.
	NPSOption func(*MusicStreamer)
)

// Ensure that a NowPlayingStreamer implements both api.MusicService and
// api.MusicStreamingService.
var (
	_ api.MusicService          = (*MusicStreamer)(nil)
	_ api.MusicStreamingService = (*MusicStreamer)(nil)
)

// WithNPSLogger adds an logger to a NowPlayingStreamer.
func WithNPSLogger(log *logrus.Logger) NPSOption {
	return func(nps *MusicStreamer) { nps.log = log }
}

// NewNowPlayingStreamer creates a new NowPlayingStreamer.
func NewNowPlayingStreamer(
	svc api.MusicService,
	interval time.Duration,
	opts ...NPSOption,
) *MusicStreamer {
	var (
		action   = func() (zero.Interface, error) { return svc.NowPlaying() }
		streamer = &MusicStreamer{
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

func (nps *MusicStreamer) startStreaming() {
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

// Stop stops the NowPlayingStreamer.
func (nps *MusicStreamer) Stop() { nps.streamer.Stop() }

// NowPlayingStream exposes a stream of NowPlaying objects.
func (nps *MusicStreamer) NowPlayingStream() <-chan api.MaybeNowPlaying {
	return nps.stream
}

// NowPlaying returns the latest NowPlaying stream result.
func (nps *MusicStreamer) NowPlaying() (*api.NowPlaying, error) {
	nps.mux.Lock()
	defer nps.mux.Unlock()
	return nps.latest.NowPlaying, nps.latest.Err
}
