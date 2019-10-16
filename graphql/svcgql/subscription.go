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

	// Tiny state machine!
	type currentlyPlayingState uint8
	const (
		stoppedState currentlyPlayingState = iota + 1
		pausedState
		playingState
	)

	go func(
		src <-chan music.CurrentlyPlayingResult,
		dst chan<- *music.CurrentlyPlaying,
	) {
		var prevState currentlyPlayingState
		for res := range src {
			if res.HasError() {
				gengql.AddError(ctx, res.Error)
				continue
			}

			if curr := res.Current; curr != nil {
				// Don't update if currently paused and previously paused.
				if (curr.Playing == false) && (prevState == pausedState) {
					goto State
				}
			} else if prevState == stoppedState {
				// Don't update if currently nil and previously stopped.
				goto State
			}

			// Send update.
			dst <- res.Current

			// Update prevState.
		State:
			if curr := res.Current; curr != nil {
				if curr.Playing {
					prevState = playingState
				} else {
					prevState = pausedState
				}
			} else {
				prevState = stoppedState
			}
		}
		close(dst)
	}(src, dst)
	if err := res.music.StreamCurrent(ctx, src); err != nil {
		return nil, err
	}
	return dst, nil
}