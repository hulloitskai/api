package svcgql

import (
	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/api/v2/assist/transit/transgql"

	"go.stevenxie.me/api/v2/about"
	"go.stevenxie.me/api/v2/about/aboutgql"

	"go.stevenxie.me/api/v2/auth"
	"go.stevenxie.me/api/v2/git"
	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/location"
	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/productivity"
	"go.stevenxie.me/api/v2/scheduling"
)

// NewResolverRoot creates a new graphql.ResolverRoot
func NewResolverRoot(svcs Services, strms Streamers) graphql.ResolverRoot {
	return resolverRoot{
		query:        newQueryResolver(svcs),
		mutation:     newMutationResolver(svcs),
		subscription: newSubscriptionResolver(strms),

		musicResolvers:        newMusicResolvers(svcs.Music),
		locationResolvers:     locationResolvers{},
		productivityResolvers: productivityResolvers{},

		fullAbout:        aboutgql.Resolver{},
		transitDeparture: transgql.DepartureResolver{},
	}
}

type (
	resolverRoot struct {
		query        graphql.QueryResolver
		mutation     graphql.MutationResolver
		subscription graphql.SubscriptionResolver

		*musicResolvers
		locationResolvers
		productivityResolvers

		fullAbout        graphql.FullAboutResolver
		transitDeparture graphql.TransitDepartureResolver
	}

	// Services handles requests for a graphql.ResolverRoot.
	Services struct {
		Git          git.Service
		Auth         auth.Service
		About        about.Service
		Music        music.Service
		Transit      transit.Service
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

func (root resolverRoot) TransitDeparture() graphql.TransitDepartureResolver {
	return root.transitDeparture
}
