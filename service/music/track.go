package music

import "github.com/zmb3/spotify"

// A Track is a unit of music.
type Track struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Artists  []*Artist `json:"artists"`
	Album    *Album    `json:"album"`
	Duration int       `json:"duration"`
}

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
