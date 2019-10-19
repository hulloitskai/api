package music

import (
	"context"
	"time"
)

// CurrentlyPlaying describes the track that is currently playing.
type CurrentlyPlaying struct {
	Timestamp time.Time     `json:"timestamp"`
	Track     Track         `json:"track"`
	Progress  time.Duration `json:"progress"`
	Playing   bool          `json:"playing"`
}

// A CurrentSource can get CurrentlyPlaying information.
type CurrentSource interface {
	GetCurrent(ctx context.Context) (*CurrentlyPlaying, error)
}

type (
	// A CurrentService handles requests for a description of my currently
	// playing music.
	CurrentService interface {
		Service()
		CurrentSource
	}

	// A CurrentStreamer can stream descriptions of my currently playing music.
	CurrentStreamer interface {
		StreamCurrent(ctx context.Context, ch chan<- CurrentlyPlayingResult) error
	}
)

// A CurrentlyPlayingResult is the result of a request for currently
// playing music descriptions.
type CurrentlyPlayingResult struct {
	Current *CurrentlyPlaying
	Error   error
}

// HasError returns true if the CurrentlyPlayingResult has an error.
func (res CurrentlyPlayingResult) HasError() bool {
	return res.Error != nil
}
