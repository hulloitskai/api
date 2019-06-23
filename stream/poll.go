package stream

import (
	"time"

	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A PollStreamer implements streaming capabilities by polling an external
	// endpoint.
	PollStreamer struct {
		poll     PollFunc
		interval time.Duration

		// Channel of either poll results or errors.
		stream chan zero.Interface

		// Channel that stops the polling goroutine.
		done chan zero.Struct
	}

	// A PollFunc polls an endpoint and returns a result and an optional error.
	PollFunc func() (zero.Interface, error)
)

// NewPollStreamer creates a new PollStreamer, which basically runs a particular
// action at a set interval, and exposes the results as a stream.
func NewPollStreamer(action PollFunc, interval time.Duration) *PollStreamer {
	ps := &PollStreamer{
		poll:     action,
		interval: interval,
		stream:   make(chan zero.Interface),
		done:     make(chan zero.Struct),
	}
	go ps.startPolling()
	return ps
}

func (ps *PollStreamer) startPolling() {
	// Start ticker.
	ticker := time.NewTicker(ps.interval)

	// Perform initial action run.
	res, err := ps.poll()
	if err != nil {
		ps.stream <- err
	} else {
		ps.stream <- res
	}

	for {
		select {
		case <-ps.done:
			ticker.Stop()
			close(ps.stream)
			return

		case <-ticker.C:
			res, err := ps.poll()
			if err != nil {
				ps.stream <- err
			} else {
				ps.stream <- res
			}
		}
	}
}

// Stop stops the PollStreamer.
//
// It can be called multiple times with no side effects.
func (ps *PollStreamer) Stop() {
	ps.done <- zero.Empty // no need to close
}

// Stream exposes PollStreamer's poll result stream.
//
// The objects returned in the stream are either of type error, or the
// streamer's PollFunc result type.
func (ps *PollStreamer) Stream() <-chan zero.Interface { return ps.stream }
