package svcgql

import (
	"context"

	"go.stevenxie.me/api/v2/auth"
	"go.stevenxie.me/api/v2/auth/authutil"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/music/musicgql"
)

func newMutationResolver(svcs Services) graphql.MutationResolver {
	return mutationResolver{
		music: musicgql.NewMutation(svcs.Music),
		auth:  svcs.Auth,
	}
}

type mutationResolver struct {
	music musicgql.Mutation
	auth  auth.Service
}

var _ graphql.MutationResolver = (*mutationResolver)(nil)

func (res mutationResolver) Music(
	ctx context.Context,
	code string,
) (*musicgql.Mutation, error) {
	ok, err := res.auth.HasPermission(ctx, code, music.PermControl)
	if err != nil {
		return nil, errors.Wrap(err, "svcgql: checking permissions")
	}
	if !ok {
		return nil, authutil.ErrAccessDenied
	}
	return &res.music, nil
}
