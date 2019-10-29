package musicsvc

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/pkg/poll"
	"go.stevenxie.me/gopkg/logutil"
)

// NewCurrentStreamer creates a new CurrentService.
func NewCurrentStreamer(
	curr music.CurrentService,
	opts ...CurrentStreamerOption,
) CurrentStreamer {
	opt := CurrentStreamerOptions{
		Logger: logutil.NoopEntry(),
	}
	for _, apply := range opts {
		apply(&opt)
	}
	log := logutil.WithComponent(opt.Logger, (*CurrentStreamer)(nil))
	var (
		actor  = newCurrentStreamActor(curr, log)
		poller = poll.NewPoller(
			actor, opt.PollInterval,
			poll.PollerWithLogger(log),
		)
	)
	return CurrentStreamer{
		curr: curr,
		poll: poller,
		act:  actor,
		log:  log,
	}
}

// StreamerWithLogger configures a CurrentStreamer to write logs with
// log.
func StreamerWithLogger(log *logrus.Entry) CurrentStreamerOption {
	return func(opt *CurrentStreamerOptions) { opt.Logger = log }
}

// StreamerWithPollInterval configures the interval at which a
// CurrentStreamer polls for changes.
func StreamerWithPollInterval(interval time.Duration) CurrentStreamerOption {
	return func(opt *CurrentStreamerOptions) { opt.PollInterval = interval }
}

type (
	// A CurrentStreamer can stream information about my currently playing music.
	CurrentStreamer struct {
		curr music.CurrentService
		poll *poll.Poller
		act  *currentStreamActor
		log  *logrus.Entry
	}

	// A CurrentStreamerOptions configures a Service.
	CurrentStreamerOptions struct {
		Logger       *logrus.Entry
		PollInterval time.Duration
	}

	// A CurrentStreamerOption modifies a CurrentStreamerOptions.
	CurrentStreamerOption func(*CurrentStreamerOptions)
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
