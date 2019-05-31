package music

import (
	"time"

	"github.com/zmb3/spotify"
)

// An Artist represents an artist that plays music.
type Artist struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// An Album represents an album of tracks.
type Album struct {
	Name   string          `json:"name"`
	URL    string          `json:"url"`
	Images []spotify.Image `json:"images"`
}

// CurrentlyPlaying represents the currently playing track.
type CurrentlyPlaying struct {
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
	Progress  int       `json:"progress"`
	Duration  int       `json:"duration"`
	URL       string    `json:"url"`
	Artists   []*Artist `json:"artists"`
	Album     *Album    `json:"album"`
}

// CurrentlyPlayingService can get the currently playing track.
type CurrentlyPlayingService interface {
	CurrentlyPlaying() (*CurrentlyPlaying, error)
}
