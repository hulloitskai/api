package api

import (
	"time"

	"github.com/zmb3/spotify"
)

// NowPlaying represents music that is currently playing.
type NowPlaying struct {
	Timestamp time.Time   `json:"timestamp"`
	Playing   bool        `json:"playing"`
	Progress  int         `json:"progress"`
	Duration  int         `json:"duration"`
	Track     *MusicTrack `json:"track"`
}

// NowPlayingService can get the currently playing track.
type NowPlayingService interface{ NowPlaying() (*NowPlaying, error) }

// A MusicTrack is a unit of music.
type MusicTrack struct {
	Name    string         `json:"name"`
	URL     string         `json:"url"`
	Artists []*MusicArtist `json:"artists"`
	Album   *MusicAlbum    `json:"album"`
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
