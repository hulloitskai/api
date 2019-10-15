package musicsvc

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/poll"
	"go.stevenxie.me/gopkg/logutil"
)

// NewCurrentStreamer creates a new CurrentService.
func NewCurrentStreamer(
	curr music.CurrentService,
	opts ...CurrentStreamerOption,
) CurrentStreamer {
	cfg := CurrentStreamerConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	var (
		log   = cfg.Logger
		actor = newCurrentStreamActor(
			curr,
			logutil.WithComponent(log, "musicsvc.currentStreamActor"),
		)
		poller = poll.NewPoller(
			actor, cfg.PollInterval,
			poll.WithPollerLogger(logutil.WithComponent(log, "poll.Poller")),
		)
	)
	return CurrentStreamer{
		curr: curr,
		poll: poller,
		act:  actor,
		log:  log,
	}
}

// WithCurrentStreamerLogger configures a CurrentStreamer to write logs with
// log.
func WithCurrentStreamerLogger(log *logrus.Entry) CurrentStreamerOption {
	return func(cfg *CurrentStreamerConfig) { cfg.Logger = log }
}

// WithCurrentStreamerPollInterval configures the interval at which a
// CurrentStreamer polls for changes.
func WithCurrentStreamerPollInterval(interval time.Duration) CurrentStreamerOption {
	return func(cfg *CurrentStreamerConfig) { cfg.PollInterval = interval }
}

type (
	// A CurrentStreamer can stream information about my currently playing music.
	CurrentStreamer struct {
		curr music.CurrentSource
		poll *poll.Poller
		act  *currentStreamActor
		log  *logrus.Entry
	}

	// A CurrentStreamerConfig configures a Service.
	CurrentStreamerConfig struct {
		Logger       *logrus.Entry
		PollInterval time.Duration
	}

	// A CurrentStreamerOption modifies a ServiceConfig.
	CurrentStreamerOption func(*CurrentStreamerConfig)
)

var _ music.CurrentStreamer = (*CurrentStreamer)(nil)

// StreamCurrent implements music.CurrentStreamer.
func (stream CurrentStreamer) StreamCurrent(
	ctx context.Context,
	ch chan<- music.CurrentlyPlayingResult,
) error {
	if ch == nil {
		panic(errors.New("musicsvc: nil channel"))
	}
	stream.act.AddSub(ch)

	// Wait for context to complete, then remove actor.
	go func() {
		<-ctx.Done()
		stream.act.DelSub(ch)
	}()
	return nil
}

// Stop stops the CurrentStreamer.
func (stream CurrentStreamer) Stop() {
	stream.poll.Stop()
	stream.act.Close()
}
