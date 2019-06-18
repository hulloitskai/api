package api

import (
	"time"

	"github.com/zmb3/spotify"
)

// NowPlayingService can get the currently playing track.
type NowPlayingService interface {
	NowPlaying() (*NowPlaying, error)
}

// NowPlayingStreamingService can stream the currently playing track.
type NowPlayingStreamingService interface {
	NowPlayingStream() <-chan MaybeNowPlaying
}

// NowPlaying represents music that is currently playing.
type NowPlaying struct {
	Timestamp time.Time   `json:"timestamp"`
	Playing   bool        `json:"playing"`
	Progress  int         `json:"progress"`
	Track     *MusicTrack `json:"track"`
}

// A MusicTrack is a unit of music.
type MusicTrack struct {
	Name     string         `json:"name"`
	URL      string         `json:"url"`
	Artists  []*MusicArtist `json:"artists"`
	Album    *MusicAlbum    `json:"album"`
	Duration int            `json:"duration"`
}

// An MusicArtist represents an artist that plays music.
type MusicArtist struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// An MusicAlbum represents an album of tracks.
type MusicAlbum struct {
	Name   string          `json:"name"`
	URL    string          `json:"url"`
	Images []spotify.Image `json:"images"`
}

// MaybeNowPlaying is a value-error pair that might be a NowPlaying object, or
// it might be an error.
type MaybeNowPlaying struct {
	NowPlaying *NowPlaying
	Err        error
}

// IsError returns true if the MaybeNowPlaying is in the error state.
func (maybe *MaybeNowPlaying) IsError() bool { return maybe.Err != nil }
