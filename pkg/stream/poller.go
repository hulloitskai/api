package stream

import (
	"time"

	"go.stevenxie.me/api/pkg/zero"
)

type (
	// A Poller is a Streamer that implements streaming capabilities by polling an
	// external endpoint.
	Poller struct {
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

var _ Streamer = (*Poller)(nil)

// NewPoller creates a new Poller, which basically runs a particular
// action at a set interval, and exposes the results as a stream.
func NewPoller(action PollFunc, interval time.Duration) *Poller {
	ps := &Poller{
		poll:     action,
		interval: interval,
		stream:   make(chan zero.Interface),
		done:     make(chan zero.Struct),
	}
	go ps.run()
	return ps
}

func (ps *Poller) run() {
	// Start ticker.
	ticker := time.NewTicker(ps.interval)

	// Perform initial action run.
	res, err := ps.poll()
	if err != nil {
		ps.stream <- err
	} else {
		ps.stream <- res
	}

	// Poll every tick, break upon stop signal.
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

// Stop stops the Poller.
//
// It can be called multiple times with no side effects.
func (ps *Poller) Stop() {
	ps.done <- zero.Empty() // no need to close
}

// Stream exposes Poller's poll result stream.
//
// The objects returned in the stream are either of type error, or the
// streamer's PollFunc result type.
func (ps *Poller) Stream() <-chan zero.Interface { return ps.stream }
