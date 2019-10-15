package music

import (
	"time"

	"github.com/zmb3/spotify"
)

// A Track is a unit of playable music.
type Track struct {
	ID          string `json:"id"`
	URI         string `json:"uri"`
	ExternalURL string `json:"externalURL"`

	Name     string        `json:"name"`
	Duration time.Duration `json:"duration"`
	Album    *Album        `json:"album,omitempty"`
	Artists  []Artist      `json:"artists"`
}

// Artist represents the artist of a Track.
type Artist struct {
	ID          string `json:"id"`
	URI         string `json:"uri"`
	ExternalURL string `json:"externalURL"`

	Name string `json:"name"`
}

// An Album represents an album of Tracks.
type Album struct {
	ID          string `json:"id"`
	URI         string `json:"uri"`
	ExternalURL string `json:"externalURL"`

	Name    string   `json:"name"`
	Images  []Image  `json:"images"`
	Artists []Artist `json:"artists"`
}

// An Image forwards a reference.
type Image spotify.Image
