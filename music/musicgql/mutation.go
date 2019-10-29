package musicgql

import (
	"context"

	"go.stevenxie.me/api/v2/music"
)

// NewMutation creates a new Mutation.
func NewMutation(svc music.Service) Mutation {
	return Mutation{svc: svc}
}

// A Mutation resolves music-related mutations.
type Mutation struct {
	svc music.Service
}

// Play plays a resource.
func (mut Mutation) Play(ctx context.Context, uri *string) (bool, error) {
	if err := mut.svc.Play(
		ctx,
		func(opt *music.PlayOptions) {
			if uri != nil {
				opt.URI = uri
			}
		},
	); err != nil {
		return false, err
	}
	return true, nil
}

// Pause pauses playback.
func (mut Mutation) Pause(ctx context.Context) (bool, error) {
	if err := mut.svc.Pause(ctx); err != nil {
		return false, err
	}
	return true, nil
}
