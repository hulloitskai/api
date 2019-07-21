package stream

import "github.com/stevenxie/api/pkg/zero"

// A Streamer can stream arbitrary objects.
type Streamer interface {
	// Stream returns a channel of objects.
	Stream() <-chan zero.Interface

	// Stop stops the Streamer.
	//
	// It can be called multiple times with no side effects.
	Stop()
}
