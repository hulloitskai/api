package svcgql

import (
	"context"

	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/gopkg/zero"
)

func newMusicAlbumResolver(svc music.SourceService) graphql.MusicAlbumResolver {
	return musicAlbumResolver{svc: svc}
}

type musicAlbumResolver struct {
	svc music.SourceService
}

var _ graphql.MusicAlbumResolver = (*musicAlbumResolver)(nil)

func (res musicAlbumResolver) Tracks(
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

func newMusicTrackResolver(svc music.SourceService) graphql.MusicTrackResolver {
	return musicTrackResolver{svc: svc}
}

type musicTrackResolver struct {
	svc music.SourceService
}

var _ graphql.MusicTrackResolver = (*musicTrackResolver)(nil)

func (res musicTrackResolver) Album(
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

func (res musicTrackResolver) Duration(
	_ context.Context,
	t *music.Track,
) (int, error) {
	return int(t.Duration.Milliseconds()), nil
}

func newMusicArtistResolver(svc music.SourceService) graphql.MusicArtistResolver {
	return musicArtistResolver{svc: svc}
}

type musicArtistResolver struct {
	svc music.SourceService
}

var _ graphql.MusicArtistResolver = (*musicArtistResolver)(nil)

func (res musicArtistResolver) Albums(
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

func newCurrentlyPlayingMusicResolver() graphql.CurrentlyPlayingMusicResolver {
	return currentlyPlayingResolver{}
}

type currentlyPlayingResolver zero.Struct

var _ graphql.CurrentlyPlayingMusicResolver = (*currentlyPlayingResolver)(nil)

func (res currentlyPlayingResolver) Progress(
	_ context.Context,
	cp *music.CurrentlyPlaying,
) (int, error) {
	return int(cp.Progress.Microseconds()), nil
}
