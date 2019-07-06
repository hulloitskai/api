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
	NowPlayingStream() <-chan MaybeNowPlaying
}

// NowPlaying represents music that is currently playing.
type NowPlaying struct {
	Timestamp time.Time    `json:"timestamp"`
	Playing   bool         `json:"playing"`
	Progress  int          `json:"progress"`
	Track     *music.Track `json:"track"`
}

// MaybeNowPlaying is a value-error pair that might be a NowPlaying object, or
// it might be an error.
type MaybeNowPlaying struct {
	NowPlaying *NowPlaying
	Err        error
}

// IsError returns true if the MaybeNowPlaying is in the error state.
func (maybe *MaybeNowPlaying) IsError() bool { return maybe.Err != nil }
