package musicgql

import (
	"context"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/gopkg/zero"
)

// NewTrackResolver creates a new TrackResolver.
func NewTrackResolver(svc music.SourceService) TrackResolver {
	return TrackResolver{svc: svc}
}

// A TrackResolver resolves fields for a music.Track.
type TrackResolver struct {
	svc music.SourceService
}

//revive:disable-line:exported
func (res TrackResolver) Album(
	ctx context.Context,
	t *music.Track,
) (*music.Album, error) {
	if t.Album == nil {
		var err error
		if t, err = res.svc.GetTrack(ctx, t.ID); err != nil {
			return nil, err
		}
	}
	return t.Album, nil
}

//revive:disable-line:exported
func (TrackResolver) Duration(_ context.Context, t *music.Track) (int, error) {
	return int(t.Duration.Milliseconds()), nil
}

// NewAlbumResolver creates a new AlbumResolver.
func NewAlbumResolver(svc music.SourceService) AlbumResolver {
	return AlbumResolver{svc: svc}
}

// An AlbumResolver resolves fields for a music.Album.
type AlbumResolver struct {
	svc music.SourceService
}

//revive:disable-line:exported
func (res AlbumResolver) Tracks(
	ctx context.Context,
	a *music.Album,
	limit, offset *int,
) ([]music.Track, error) {
	return res.svc.GetAlbumTracks(
		ctx,
		a.ID,
		func(cfg *music.PaginationConfig) {
			if limit != nil {
				cfg.Limit = *limit
			}
			if offset != nil {
				cfg.Offset = *offset
			}
		},
	)
}

// NewArtistResolver creates a new ArtistResolver.
func NewArtistResolver(svc music.SourceService) ArtistResolver {
	return ArtistResolver{svc: svc}
}

// An ArtistResolver resolves fields for a music.Artist.
type ArtistResolver struct {
	svc music.SourceService
}

//revive:disable-line:exported
func (res ArtistResolver) Albums(
	ctx context.Context,
	t *music.Artist,
	limit, offset *int,
) ([]music.Album, error) {
	return res.svc.GetArtistAlbums(
		ctx,
		t.ID,
		func(cfg *music.PaginationConfig) {
			if limit != nil {
				cfg.Limit = *limit
			}
			if offset != nil {
				cfg.Offset = *offset
			}
		},
	)
}

// A CurrentlyPlayingResolver resolves fields for a music.CurrentlyPlaying.
type CurrentlyPlayingResolver zero.Struct

//revive:disable-line:exported
func (CurrentlyPlayingResolver) Progress(
	_ context.Context,
	cp *music.CurrentlyPlaying,
) (int, error) {
	return int(cp.Progress.Milliseconds()), nil
}
