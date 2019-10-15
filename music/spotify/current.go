package spotify

import (
	"context"
	"io"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/zmb3/spotify"
	"go.stevenxie.me/api/music"
)

// NewCurrentSource creates a new music.CurrentSource.
func NewCurrentSource(c *spotify.Client) music.CurrentSource {
	return currentSource{client: c}
}

type currentSource struct {
	client *spotify.Client
}

var _ music.CurrentSource = (*currentSource)(nil)

func (src currentSource) GetCurrent(context.Context) (
	*music.CurrentlyPlaying,
	error,
) {
	cp, err := src.client.PlayerCurrentlyPlaying()
	if err != nil {
		if errors.Is(err, io.EOF) { // nothing is playing
			return nil, nil
		}
		return nil, err
	}
	item := cp.Item

	// Parse timestamp as time.Time.
	timestamp := time.Unix(
		cp.Timestamp/1000,
		(cp.Timestamp%1000)*int64((time.Millisecond/time.Nanosecond)),
	)

	// Parse track.
	track := music.Track{Album: new(music.Album)}
	trackFromSpotify(&track, &item.SimpleTrack)
	albumFromSpotify(track.Album, &item.Album)

	return &music.CurrentlyPlaying{
		Timestamp: timestamp,
		Playing:   cp.Playing,
		Progress:  time.Duration(cp.Progress) * time.Millisecond,
		Track:     track,
	}, nil
}
