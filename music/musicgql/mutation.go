package musicgql

import (
	"context"

	"go.stevenxie.me/api/music"
)

// NewMutation creates a new Mutation.
func NewMutation(svc music.Service) Mutation {
	return Mutation{svc: svc}
}

// A Mutation resolves queries for my music-related data.
type Mutation struct {
	svc music.Service
}

// Play plays a resource.
func (q Mutation) Play(ctx context.Context, uri *string) (bool, error) {
	if err := q.svc.Play(
		ctx,
		func(cfg *music.PlayConfig) { cfg.URI = uri },
	); err != nil {
		return false, err
	}
	return true, nil
}

// Pause pauses playback.
func (q Mutation) Pause(ctx context.Context) (bool, error) {
	if err := q.svc.Pause(ctx); err != nil {
		return false, err
	}
	return true, nil
}
