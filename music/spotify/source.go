package spotify

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/zmb3/spotify"
	"go.stevenxie.me/api/music"
)

// NewSource creates a new music.Source.
func NewSource(c *spotify.Client) music.Source {
	return source{client: c}
}

type source struct {
	client *spotify.Client
}

var _ music.Source = (*source)(nil)

func (src source) GetTrack(_ context.Context, id string) (*music.Track, error) {
	st, err := src.client.GetTrack(spotify.ID(id))
	if err != nil {
		return nil, errors.WithMessage(err, "spotify")
	}

	t := &music.Track{Album: new(music.Album)}
	trackFromSpotify(t, &st.SimpleTrack)
	albumFromSpotify(t.Album, &st.Album)
	return t, nil
}

func (src source) GetAlbumTracks(
	_ context.Context,
	id string,
	limit, offset int,
) ([]music.Track, error) {
	sts, err := src.client.GetAlbumTracksOpt(spotify.ID(id), limit, offset)
	if err != nil {
		return nil, errors.WithMessage(err, "spotify")
	}
	var ts []music.Track
	tracksFromSpotify(&ts, sts.Tracks)
	return ts, nil
}

func (src source) GetArtistAlbums(
	_ context.Context,
	id string,
	limit, offset int,
) ([]music.Album, error) {
	sas, err := src.client.GetArtistAlbumsOpt(
		spotify.ID(id),
		&spotify.Options{
			Limit:  &limit,
			Offset: &offset,
		},
		nil,
	)
	if err != nil {
		return nil, errors.WithMessage(err, "spotify")
	}
	var as []music.Album
	albumsFromSpotify(&as, sas.Albums)
	return as, nil
}
