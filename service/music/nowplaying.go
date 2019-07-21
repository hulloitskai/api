package music

import "time"

// NowPlayingService can get information about the music being played.
type NowPlayingService interface {
	NowPlaying() (*NowPlaying, error)
}

// NowPlayingStreamingService can get and stream information about the music
// being played.
type NowPlayingStreamingService interface {
	NowPlayingService
	NowPlayingStream() <-chan struct {
		NowPlaying *NowPlaying
		Err        error
	}
}

// NowPlaying represents music that is currently playing.
type NowPlaying struct {
	Timestamp time.Time `json:"timestamp"`
	Playing   bool      `json:"playing"`
	Progress  int       `json:"progress"`
	Track     *Track    `json:"track"`
}
