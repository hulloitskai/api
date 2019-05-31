package spotify

import (
	"time"

	"github.com/stevenxie/api/pkg/spotify"
)

const extURLKeySpotify = "spotify"

// GetCurrentlyPlaying returns the currently playing track.
func (c *Client) GetCurrentlyPlaying() (*spotify.CurrentlyPlaying, error) {
	cp, err := c.sc.PlayerCurrentlyPlaying()
	if err != nil {
		return nil, err
	}

	// Derive track album.
	var (
		sa    = &cp.Item.Album
		album = &spotify.Album{
			Name:   sa.Name,
			URL:    extSpotifyURL(sa.ExternalURLs),
			Images: sa.Images,
		}
	)

	// Derive track artists.
	artists := make([]*spotify.Artist, len(cp.Item.Artists))
	for i, a := range cp.Item.Artists {
		artists[i] = &spotify.Artist{
			Name: a.Name,
			URL:  extSpotifyURL(a.ExternalURLs),
		}
	}

	return &spotify.CurrentlyPlaying{
		Name: cp.Item.Name,
		Timestamp: time.Unix(
			cp.Timestamp/1000,
			(cp.Timestamp%1000)*int64((time.Millisecond/time.Nanosecond)),
		),
		Progress: cp.Progress,
		Duration: cp.Item.Duration,
		URL:      extSpotifyURL(cp.Item.ExternalURLs),
		Artists:  artists,
		Album:    album,
	}, nil
}

func extSpotifyURL(urls map[string]string) string {
	return urls[extURLKeySpotify]
}
