package api

import (
	"time"

	"github.com/stevenxie/api/pkg/music"
)

// MusicService can get information about the music being played.
type MusicService interface {
	NowPlaying() (*NowPlaying, error)
}

// MusicStreamingService can stream information about the music being played.
type MusicStreamingService interface {
	MusicService
	NowPlayingStream() <-chan struct {
		NowPlaying *NowPlaying
		Err        error
	}
}

// NowPlaying represents music that is currently playing.
type NowPlaying struct {
	Timestamp time.Time    `json:"timestamp"`
	Playing   bool         `json:"playing"`
	Progress  int          `json:"progress"`
	Track     *music.Track `json:"track"`
}
