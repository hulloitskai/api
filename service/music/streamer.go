package music

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/stream"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// NowPlayingStreamer implements a streaming interface over a
	// NowPlayingService.
	NowPlayingStreamer struct {
		streamer stream.Streamer
		log      logrus.FieldLogger

		stream chan struct {
			NowPlaying *NowPlaying
			Err        error
		}

		mux        sync.Mutex
		nowPlaying *NowPlaying
		err        error
	}

	// An NowPlayingStreamerConfig configures a NowPlayingStreamer.
	NowPlayingStreamerConfig struct {
		Logger logrus.FieldLogger
	}
)

var _ NowPlayingStreamingService = (*NowPlayingStreamer)(nil)

// NewNowPlayingStreamer creates a new NowPlayingStreamer.
func NewNowPlayingStreamer(
	svc NowPlayingService,
	interval time.Duration,
	opts ...func(*NowPlayingStreamerConfig),
) *NowPlayingStreamer {
	cfg := NowPlayingStreamerConfig{Logger: zero.Logger()}
	for _, opt := range opts {
		opt(&cfg)
	}

	nps := &NowPlayingStreamer{
		streamer: stream.NewPoller(
			func() (zero.Interface, error) { return svc.NowPlaying() },
			interval,
		),
		stream: make(chan struct {
			NowPlaying *NowPlaying
			Err        error
		}),
		log: zero.Logger(),
	}
	go nps.run()
	return nps
}

func (nps *NowPlayingStreamer) run() {
	for result := range nps.streamer.Stream() {
		var (
			nowPlaying *NowPlaying
			err        error
		)

		switch v := result.(type) {
		case error:
			err = v
			nps.log.WithError(err).Error("Failed to load now-playing data.")
		case *NowPlaying:
			nowPlaying = v
		default:
			nps.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("music: unexpected upstream value '%v'", v)
		}

		// Cache values.
		nps.mux.Lock()
		nps.nowPlaying = nowPlaying
		nps.err = err
		nps.mux.Unlock()

		// Write values to stream.
		nps.stream <- struct {
			NowPlaying *NowPlaying
			Err        error
		}{nowPlaying, err}
	}
}

// Stop stops the NowPlayingStreamer.
func (nps *NowPlayingStreamer) Stop() { nps.streamer.Stop() }

// NowPlayingStream returns a stream of NowPlaying objects.
func (nps *NowPlayingStreamer) NowPlayingStream() <-chan struct {
	NowPlaying *NowPlaying
	Err        error
} {
	return nps.stream
}

// NowPlaying returns the latest NowPlaying stream result.
func (nps *NowPlayingStreamer) NowPlaying() (*NowPlaying, error) {
	nps.mux.Lock()
	defer nps.mux.Unlock()
	return nps.nowPlaying, nps.err
}
