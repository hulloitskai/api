package svcgql

import (
	"context"

	gengql "github.com/99designs/gqlgen/graphql"

	"go.stevenxie.me/api/graphql"
	"go.stevenxie.me/api/music"
)

func newSubscriptionResolver(strms Streamers) graphql.SubscriptionResolver {
	return subscriptionResolver{
		music: strms.Music,
	}
}

type subscriptionResolver struct {
	music music.Streamer
}

var _ graphql.SubscriptionResolver = (*subscriptionResolver)(nil)

func (res subscriptionResolver) Music(ctx context.Context) (
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
		for res := range src {
			if res.HasError() {
				gengql.AddError(ctx, res.Error)
				continue
			}
			dst <- res.Current
		}
		close(dst)
	}(src, dst)
	if err := res.music.StreamCurrent(ctx, src); err != nil {
		return nil, err
	}
	return dst, nil
}
