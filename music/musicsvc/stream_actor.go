package musicsvc

import (
	"context"
	"sync"

	"go.stevenxie.me/gopkg/logutil"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/poll"
	"go.stevenxie.me/gopkg/zero"
)

func newCurrentStreamActor(
	src music.CurrentService,
	log *logrus.Entry,
) *currentStreamActor {
	return &currentStreamActor{
		svc:  src,
		log:  logutil.AddComponent(log, (*currentStreamActor)(nil)),
		subs: make(map[chan<- music.CurrentlyPlayingResult]zero.Struct),
	}
}

type currentStreamActor struct {
	svc music.CurrentService
	log *logrus.Entry

	mux  sync.Mutex
	subs map[chan<- music.CurrentlyPlayingResult]zero.Struct
}

var _ poll.Actor = (*currentStreamActor)(nil)

func (act *currentStreamActor) Prod() (zero.Interface, error) {
	return act.svc.GetCurrent(context.Background())
}

func (act *currentStreamActor) Recv(v zero.Interface, err error) {
	log := logutil.WithMethod(act.log, (*currentStreamActor).Recv)

	// Parse received value.
	cp, ok := v.(*music.CurrentlyPlaying)
	if !ok {
		log.
			WithError(err).
			WithField("value", v).
			Error("Received an unknown value.")
		panic(errors.Newf("musicsvc: actor received unknown value '%T'", v))
	}

	// Send to all subscribers.
	act.mux.Lock()
	if act.subs != nil {
		for ch := range act.subs {
			ch <- music.CurrentlyPlayingResult{
				Current: cp,
				Error:   err,
			}
		}
	}
	act.mux.Unlock()
}

// Add listener channel. Is concurrent-safe.
func (act *currentStreamActor) AddSub(
	ch chan<- music.CurrentlyPlayingResult,
) {
	act.mux.Lock()
	if act.subs != nil {
		act.subs[ch] = zero.Empty()
	}
	act.mux.Unlock()
}

// Delete listener channel. Is concurrent-safe.
func (act *currentStreamActor) DelSub(
	ch chan<- music.CurrentlyPlayingResult,
) {
	act.mux.Lock()
	if act.subs != nil {
		delete(act.subs, ch)
	}
	act.mux.Unlock()
}

// Close closes all channels, free subs.
func (act *currentStreamActor) Close() {
	act.mux.Lock()
	for ch := range act.subs {
		close(ch)
	}
	act.subs = nil
	act.mux.Unlock()
}
