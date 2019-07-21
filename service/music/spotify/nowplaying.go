package spotify

import (
	"io"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/service/music"
	"github.com/zmb3/spotify"
)

const extURLKeySpotify = "spotify"

// A NowPlayingService can access the Spotify API.
type NowPlayingService struct {
	client *spotify.Client
}

var _ music.NowPlayingService = (*NowPlayingService)(nil)

// NewNowPlayingService creates a new NowPlayingService.
func NewNowPlayingService(c *spotify.Client) NowPlayingService {
	return NowPlayingService{client: c}
}

// NowPlaying returns the currently playing track.
func (svc NowPlayingService) NowPlaying() (*music.NowPlaying, error) {
	cp, err := svc.client.PlayerCurrentlyPlaying()
	if err != nil {
		if errors.Is(err, io.EOF) { // if EOF, return nill (no current playing song)
			return nil, nil
		}
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

	return &music.NowPlaying{
		Timestamp: time.Unix(
			cp.Timestamp/1000,
			(cp.Timestamp%1000)*int64((time.Millisecond/time.Nanosecond)),
		),
		Playing:  cp.Playing,
		Progress: cp.Progress,
		Track: &music.Track{
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
