package stream

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
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

		stream chan struct {
			NowPlaying *api.NowPlaying
			Err        error
		}

		mux        sync.Mutex
		nowPlaying *api.NowPlaying
		err        error
	}

	// An MSConfig configures a MusicStreamer.
	MSConfig struct{ Logger *logrus.Logger }
)

// Ensure that a MusicStreamer implements both api.MusicService and
// api.MusicStreamingService.
var (
	_ api.MusicService          = (*MusicStreamer)(nil)
	_ api.MusicStreamingService = (*MusicStreamer)(nil)
)

// NewMusicStreamer creates a new MusicStreamer.
func NewMusicStreamer(
	svc api.MusicService,
	interval time.Duration,
	opts ...func(*MSConfig),
) *MusicStreamer {
	cfg := MSConfig{Logger: zero.Logger()}
	for _, opt := range opts {
		opt(&cfg)
	}

	ms := &MusicStreamer{
		streamer: NewPollStreamer(
			func() (zero.Interface, error) { return svc.NowPlaying() },
			interval,
		),
		stream: make(chan struct {
			NowPlaying *api.NowPlaying
			Err        error
		}),
		log: zero.Logger(),
	}
	go ms.startStreaming()
	return ms
}

func (ms *MusicStreamer) startStreaming() {
	for result := range ms.streamer.Stream() {
		var (
			nowPlaying *api.NowPlaying
			err        error
		)

		switch v := result.(type) {
		case error:
			err = v
			ms.log.WithError(err).Error("Failed to load now-playing data.")
		case *api.NowPlaying:
			nowPlaying = v
		default:
			ms.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("stream: unexpected upstream value '%v'", v)
		}

		// Cache values.
		ms.mux.Lock()
		ms.nowPlaying = nowPlaying
		ms.err = err
		ms.mux.Unlock()

		// Write values to stream.
		ms.stream <- struct {
			NowPlaying *api.NowPlaying
			Err        error
		}{nowPlaying, err}
	}
}

// Stop stops the MusicStreamer.
func (ms *MusicStreamer) Stop() { ms.streamer.Stop() }

// NowPlayingStream returns a stream of NowPlaying objects.
func (ms *MusicStreamer) NowPlayingStream() <-chan struct {
	NowPlaying *api.NowPlaying
	Err        error
} {
	return ms.stream
}

// NowPlaying returns the latest NowPlaying stream result.
func (ms *MusicStreamer) NowPlaying() (*api.NowPlaying, error) {
	ms.mux.Lock()
	defer ms.mux.Unlock()
	return ms.nowPlaying, ms.err
}
