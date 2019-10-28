package musicgql

import (
	"context"

	"go.stevenxie.me/api/v2/music"
)

// NewQuery creates a new Query.
func NewQuery(svc music.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for my music-related data.
type Query struct {
	svc music.Service
}

// Current gets my current playing music information.
func (q Query) Current(ctx context.Context) (*music.CurrentlyPlaying, error) {
	return q.svc.GetCurrent(ctx)
}
