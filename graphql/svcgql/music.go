package svcgql

import (
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/music/musicgql"
)

func newMusicResolvers(svc music.Service) *musicResolvers {
	return &musicResolvers{
		track:   musicgql.NewTrackResolver(svc),
		album:   musicgql.NewAlbumResolver(svc),
		artist:  musicgql.NewArtistResolver(svc),
		current: musicgql.CurrentlyPlayingResolver{},
	}
}

type musicResolvers struct {
	track   musicgql.TrackResolver
	album   musicgql.AlbumResolver
	artist  musicgql.ArtistResolver
	current musicgql.CurrentlyPlayingResolver
}

func (res *musicResolvers) MusicTrack() graphql.MusicTrackResolver   { return res.track }
func (res *musicResolvers) MusicAlbum() graphql.MusicAlbumResolver   { return res.album }
func (res *musicResolvers) MusicArtist() graphql.MusicArtistResolver { return res.artist }
func (res *musicResolvers) CurrentlyPlayingMusic() graphql.CurrentlyPlayingMusicResolver {
	return res.current
}
