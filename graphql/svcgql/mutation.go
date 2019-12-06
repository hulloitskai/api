package svcgql

import (
	"context"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/v2/auth/authutil"
	"go.stevenxie.me/api/v2/graphql"
	"go.stevenxie.me/api/v2/music"
)

func newMutationResolver(svcs Services) graphql.MutationResolver {
	return mutationResolver{
		svcs: svcs,
	}
}

type mutationResolver struct {
	svcs Services
}

var _ graphql.MutationResolver = (*mutationResolver)(nil)

func (res mutationResolver) PlayMusic(
	ctx context.Context,
	code string,
	resource *music.Selector,
) (bool, error) {
	if err := res.checkMusicCode(ctx, code); err != nil {
		return false, err
	}
	if err := res.svcs.Music.Play(
		ctx,
		func(opt *music.PlayOptions) {
			if resource != nil {
				opt.Selector = resource
			}
		},
	); err != nil {
		return false, err
	}
	return true, nil
}

func (res mutationResolver) PauseMusic(
	ctx context.Context,
	code string,
) (bool, error) {
	if err := res.checkMusicCode(ctx, code); err != nil {
		return false, err
	}
	if err := res.svcs.Music.Pause(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (res mutationResolver) checkMusicCode(
	ctx context.Context,
	code string,
) error {
	ok, err := res.svcs.Auth.HasPermission(ctx, code, music.PermControl)
	if err != nil {
		return errors.Wrap(err, "svcgql: check permissions")
	}
	if !ok {
		return authutil.ErrAccessDenied
	}
	return nil
}
