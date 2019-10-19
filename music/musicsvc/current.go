package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewCurrentService creates a new CurrentService.
func NewCurrentService(
	src music.CurrentSource,
	opts ...svcutil.BasicOption,
) music.CurrentService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return currentService{
		curr: src,
		log:  logutil.AddComponent(cfg.Logger, (*currentService)(nil)),
	}
}

type currentService struct {
	curr music.CurrentSource
	log  *logrus.Entry
}

var _ music.CurrentService = (*currentService)(nil)

func (svc currentService) GetCurrent(ctx context.Context) (
	*music.CurrentlyPlaying,
	error,
) {
	log := logutil.
		WithMethod(svc.log, currentService.GetCurrent).
		WithContext(ctx)

	cp, err := svc.curr.GetCurrent(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get currently playing music.")
		return nil, err
	}
	log.WithField("current", cp).Trace("Got currently playing music.")

	return cp, nil
}

func (currentService) Service() {}
