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

	// An MSOption configures a MusicStreamer.
	MSOption func(*MusicStreamer)
)

// Ensure that a MusicStreamer implements both api.MusicService and
// api.MusicStreamingService.
var (
	_ api.MusicService          = (*MusicStreamer)(nil)
	_ api.MusicStreamingService = (*MusicStreamer)(nil)
)

// WithMSLogger adds an logger to a MusicStreamer.
func WithMSLogger(log *logrus.Logger) MSOption {
	return func(ms *MusicStreamer) { ms.log = log }
}

// NewMusicStreamer creates a new MusicStreamer.
func NewMusicStreamer(
	svc api.MusicService,
	interval time.Duration,
	opts ...MSOption,
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

func (ms *MusicStreamer) startStreaming() {
	for result := range ms.streamer.Stream() {
		var maybe api.MaybeNowPlaying
		switch v := result.(type) {
		case error:
			maybe = api.MaybeNowPlaying{Err: v}
			ms.log.WithError(maybe.Err).Error("Failed to load now-playing data.")
		case *api.NowPlaying:
			maybe = api.MaybeNowPlaying{NowPlaying: v}
		default:
			ms.log.WithField("value", v).Error("Unexpected value from upstream.")
			maybe = api.MaybeNowPlaying{
				Err: errors.Errorf("stream: unexpected upstream value (%v)", v),
			}
		}

		// Safely write maybe to latest.
		ms.mux.Lock()
		ms.latest = maybe
		ms.mux.Unlock()

		// Write maybe to stream.
		ms.stream <- maybe
	}
}

// Stop stops the MusicStreamer.
func (ms *MusicStreamer) Stop() { ms.streamer.Stop() }

// NowPlayingStream exposes a stream of NowPlaying objects.
func (ms *MusicStreamer) NowPlayingStream() <-chan api.MaybeNowPlaying {
	return ms.stream
}

// NowPlaying returns the latest NowPlaying stream result.
func (ms *MusicStreamer) NowPlaying() (*api.NowPlaying, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.latest.NowPlaying, ms.latest.Err
}
