package spotify

import (
	"io"
	"time"

	"github.com/stevenxie/api/pkg/api"
	errors "golang.org/x/xerrors"
)

const extURLKeySpotify = "spotify"

// NowPlaying returns the currently playing track.
func (c *Client) NowPlaying() (*api.NowPlaying, error) {
	cp, err := c.sc.PlayerCurrentlyPlaying()
	if err != nil {
		if errors.Is(err, io.EOF) { // if EOF, return nill (no current playing song)
			return nil, nil
		}
		return nil, err
	}

	// Derive track album.
	var (
		sa    = &cp.Item.Album
		album = &api.MusicAlbum{
			Name:   sa.Name,
			URL:    extSpotifyURL(sa.ExternalURLs),
			Images: sa.Images,
		}
	)

	// Derive track artists.
	artists := make([]*api.MusicArtist, len(cp.Item.Artists))
	for i, a := range cp.Item.Artists {
		artists[i] = &api.MusicArtist{
			Name: a.Name,
			URL:  extSpotifyURL(a.ExternalURLs),
		}
	}

	return &api.NowPlaying{
		Timestamp: time.Unix(
			cp.Timestamp/1000,
			(cp.Timestamp%1000)*int64((time.Millisecond/time.Nanosecond)),
		),
		Playing:  cp.Playing,
		Progress: cp.Progress,
		Track: &api.MusicTrack{
			Name:     cp.Item.Name,
			URL:      extSpotifyURL(cp.Item.ExternalURLs),
			Artists:  artists,
			Album:    album,
			Duration: cp.Item.Duration,
		},
	}, nil
}

func extSpotifyURL(urls map[string]string) string {
	return urls[extURLKeySpotify]
}
