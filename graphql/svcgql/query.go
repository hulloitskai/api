package svcgql

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/auth"
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

		git:        gitgql.NewQuery(svcs.Git),
		music:      musicgql.NewQuery(svcs.Music),
		location:   locgql.NewQuery(svcs.Location, svcs.Auth),
		scheduling: schedgql.NewQuery(svcs.Scheduling),
	}
}

type queryResolver struct {
	prod  productivity.Service
	auth  auth.Service
	about about.Service

	git        gitgql.Query
	music      musicgql.Query
	location   locgql.Query
	scheduling schedgql.Query
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

func (qr queryResolver) Permissions(ctx context.Context, code string) (
	perms []string, err error) {
	ps, err := qr.auth.GetPermissions(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	perms = make([]string, len(ps))
	for i, p := range ps {
		perms[i] = string(p)
	}
	return perms, nil
}

func (qr queryResolver) Git(context.Context) (*gitgql.Query, error) {
	return &qr.git, nil
}

func (qr queryResolver) Music(context.Context) (*musicgql.Query, error) {
	return &qr.music, nil
}

func (qr queryResolver) Location(context.Context) (*locgql.Query, error) {
	return &qr.location, nil
}

func (qr queryResolver) Scheduling(context.Context) (*schedgql.Query, error) {
	return &qr.scheduling, nil
}
