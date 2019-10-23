package svcgql

import (
	"context"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/about/aboutgql"
	"go.stevenxie.me/api/assist/assistgql"
	"go.stevenxie.me/api/auth/authgql"
	"go.stevenxie.me/api/git/gitgql"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/location/locgql"
	"go.stevenxie.me/api/music/musicgql"
	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/productivity/prodgql"
	"go.stevenxie.me/api/scheduling/schedgql"
)

func newQueryResolver(svcs Services) graphql.QueryResolver {
	return queryResolver{
		about:  aboutgql.NewQuery(svcs.About, svcs.Auth),
		prod:   prodgql.NewQuery(svcs.Productivity),
		gitq:   gitgql.NewQuery(svcs.Git),
		locq:   locgql.NewQuery(svcs.Location, svcs.Auth),
		authq:  authgql.NewQuery(svcs.Auth),
		musicq: musicgql.NewQuery(svcs.Music),
		schedq: schedgql.NewQuery(svcs.Scheduling, svcs.Auth),
		assistq: assistgql.NewQuery(assistgql.QueryServices{
			Transit: svcs.Transit,
		}),
	}
}

type queryResolver struct {
	about aboutgql.Query
	prod  prodgql.Query

	gitq    gitgql.Query
	locq    locgql.Query
	authq   authgql.Query
	musicq  musicgql.Query
	schedq  schedgql.Query
	assistq assistgql.Query
}

var _ graphql.QueryResolver = (*queryResolver)(nil)

func (qr queryResolver) About(
	ctx context.Context,
	code *string,
) (about.ContactInfo, error) {
	return qr.about.About(ctx, code)
}

func (qr queryResolver) Productivity(ctx context.Context) (
	*productivity.Productivity, error) {
	return qr.prod.Productivity(ctx)
}

func (qr queryResolver) Git(context.Context) (*gitgql.Query, error) {
	return &qr.gitq, nil
}

func (qr queryResolver) Auth(context.Context) (*authgql.Query, error) {
	return &qr.authq, nil
}

func (qr queryResolver) Music(context.Context) (*musicgql.Query, error) {
	return &qr.musicq, nil
}

func (qr queryResolver) Assist(context.Context) (*assistgql.Query, error) {
	return &qr.assistq, nil
}

func (qr queryResolver) Location(context.Context) (*locgql.Query, error) {
	return &qr.locq, nil
}

func (qr queryResolver) Scheduling(context.Context) (*schedgql.Query, error) {
	return &qr.schedq, nil
}
