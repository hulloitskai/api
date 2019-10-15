package svcgql

import (
	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/scheduling"
)

// NewResolverRoot creates a new graphql.ResolverRoot
func NewResolverRoot(svcs Services, strms Streamers) graphql.ResolverRoot {
	return resolverRoot{
		query:        newQueryResolver(svcs),
		mutation:     newMutationResolver(svcs),
		subscription: newSubscriptionResolver(strms),

		fullAbout: newFullAboutResolver(),

		currentlyPlayingMusic: newCurrentlyPlayingMusicResolver(),
		musicAlbum:            newMusicAlbumResolver(svcs.Music),
		musicTrack:            newMusicTrackResolver(svcs.Music),
		musicArtist:           newMusicArtistResolver(svcs.Music),

		productivity:       newProductivityResolver(),
		productivityRecord: newProductivityRecordResolver(),
	}
}

type (
	resolverRoot struct {
		query        graphql.QueryResolver
		mutation     graphql.MutationResolver
		subscription graphql.SubscriptionResolver

		fullAbout graphql.FullAboutResolver

		currentlyPlayingMusic graphql.CurrentlyPlayingMusicResolver
		musicAlbum            graphql.MusicAlbumResolver
		musicTrack            graphql.MusicTrackResolver
		musicArtist           graphql.MusicArtistResolver

		productivity       graphql.ProductivityResolver
		productivityRecord graphql.ProductivityRecordResolver
	}

	// Services handles requests for a graphql.ResolverRoot.
	Services struct {
		Git          git.Service
		Auth         auth.Service
		About        about.Service
		Music        music.Service
		Location     location.Service
		Scheduling   scheduling.Service
		Productivity productivity.Service
	}

	// Streamers handles streams for a graphql.ResolverRoot.
	Streamers struct {
		Music music.Streamer
	}
)

var _ graphql.ResolverRoot = (*resolverRoot)(nil)

func (root resolverRoot) Query() graphql.QueryResolver {
	return root.query
}

func (root resolverRoot) Mutation() graphql.MutationResolver {
	return root.mutation
}

func (root resolverRoot) Subscription() graphql.SubscriptionResolver {
	return root.subscription
}

func (root resolverRoot) FullAbout() graphql.FullAboutResolver {
	return root.fullAbout
}

func (root resolverRoot) CurrentlyPlayingMusic() graphql.CurrentlyPlayingMusicResolver {
	return root.currentlyPlayingMusic
}

func (root resolverRoot) MusicAlbum() graphql.MusicAlbumResolver {
	return root.musicAlbum
}

func (root resolverRoot) MusicTrack() graphql.MusicTrackResolver {
	return root.musicTrack
}

func (root resolverRoot) MusicArtist() graphql.MusicArtistResolver {
	return root.musicArtist
}

func (root resolverRoot) Productivity() graphql.ProductivityResolver {
	return root.productivity
}

func (root resolverRoot) ProductivityRecord() graphql.ProductivityRecordResolver {
	return root.productivityRecord
}
