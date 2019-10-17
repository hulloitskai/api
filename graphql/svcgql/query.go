package svcgql

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/assist/assistgql"
	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/auth/authgql"
	"go.stevenxie.me/api/auth/authutil"
	"go.stevenxie.me/api/git/gitgql"
	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/location/locgql"
	"go.stevenxie.me/api/music/musicgql"
	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/scheduling/schedgql"
)

func newQueryResolver(svcs Services) graphql.QueryResolver {
	return queryResolver{
		about: svcs.About,
		prod:  svcs.Productivity,
		auth:  svcs.Auth,

		gitq:   gitgql.NewQuery(svcs.Git),
		locq:   locgql.NewQuery(svcs.Location, svcs.Auth),
		authq:  authgql.NewQuery(svcs.Auth),
		musicq: musicgql.NewQuery(svcs.Music),
		schedq: schedgql.NewQuery(svcs.Scheduling),
		assistq: assistgql.NewQuery(assistgql.QueryServices{
			Transit: svcs.Transit,
		}),
	}
}

type queryResolver struct {
	about about.Service
	prod  productivity.Service
	auth  auth.Service

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
	if code != nil {
		ok, err := qr.auth.HasPermission(
			ctx,
			strings.TrimSpace(*code), about.PermFull,
		)
		if err != nil {
			return nil, errors.Wrap(err, "svcgql: checking permissions")
		}
		if !ok {
			return nil, authutil.ErrAccessDenied
		}
		return qr.about.GetAbout(ctx)
	}
	return qr.about.GetMasked(ctx)
}

func (qr queryResolver) Productivity(ctx context.Context) (
	*productivity.Productivity, error) {
	return qr.prod.CurrentProductivity(ctx)
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
