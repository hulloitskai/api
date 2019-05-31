package spotify

import (
	"time"

	"github.com/stevenxie/api/pkg/music"
)

const extURLKeySpotify = "spotify"

// CurrentlyPlaying returns the currently playing track.
func (c *Client) CurrentlyPlaying() (*music.CurrentlyPlaying, error) {
	cp, err := c.sc.PlayerCurrentlyPlaying()
	if err != nil {
		return nil, err
	}

	// Derive track album.
	var (
		sa    = &cp.Item.Album
		album = &music.Album{
			Name:   sa.Name,
			URL:    extSpotifyURL(sa.ExternalURLs),
			Images: sa.Images,
		}
	)

	// Derive track artists.
	artists := make([]*music.Artist, len(cp.Item.Artists))
	for i, a := range cp.Item.Artists {
		artists[i] = &music.Artist{
			Name: a.Name,
			URL:  extSpotifyURL(a.ExternalURLs),
		}
	}

	return &music.CurrentlyPlaying{
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
