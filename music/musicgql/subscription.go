package musicgql

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"go.stevenxie.me/api/v2/music"
)

// NewSubscriptionResolver creates a new SubscriptionResolver.
func NewSubscriptionResolver(stream music.Streamer) SubscriptionResolver {
	return SubscriptionResolver{
		stream: stream,
	}
}

// A SubscriptionResolver resolves music-related GraphQL subscriptions.
type SubscriptionResolver struct {
	stream music.Streamer
}

// CurrentlyPlaying opens a music.CurrentlyPlaying stream.
func (res SubscriptionResolver) CurrentlyPlaying(ctx context.Context) (
	<-chan *music.CurrentlyPlaying,
	error,
) {
	var (
		src = make(chan music.CurrentlyPlayingResult, 1)
		dst = make(chan *music.CurrentlyPlaying, 1)
	)

	go func(
		src <-chan music.CurrentlyPlayingResult,
		dst chan<- *music.CurrentlyPlaying,
	) {
		var prev *music.CurrentlyPlaying
		for res := range src {
			if res.HasError() {
				graphql.AddError(ctx, res.Error)
				continue
			}
			curr := res.Current
			if !IsEqualsCurrentlyPlaying(prev, curr) {
				dst <- curr
			}
			prev = curr
		}
		close(dst)
	}(src, dst)

	if err := res.stream.StreamCurrent(ctx, src); err != nil {
		return nil, err
	}
	return dst, nil
}
