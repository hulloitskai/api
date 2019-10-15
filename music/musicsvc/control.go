package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
)

// NewControlService creates a new music.ControlService.
func NewControlService(
	ctrl music.Controller,
	opts ...svcutil.BasicOption,
) music.ControlService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return controlService{
		ctrl: ctrl,
		log:  cfg.Logger,
	}
}

type controlService struct {
	ctrl music.Controller
	log  *logrus.Entry
}

var _ music.ControlService = (*controlService)(nil)

func (svc controlService) Play(
	ctx context.Context,
	opts ...music.PlayOption,
) error {
	var cfg music.PlayConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	log := svc.log.WithFields(logrus.Fields{
		"method": "Play",
		"uri":    cfg.URI,
	}).WithContext(ctx)
	if err := svc.ctrl.Play(ctx, cfg.URI); err != nil {
		log.WithError(err).Error("Failed to play resource.")
		return err
	}
	log.Trace("Started playing resource.")
	return nil
}

func (svc controlService) Pause(ctx context.Context) error {
	log := svc.log.
		WithField("method", "Pause").
		WithContext(ctx)
	if err := svc.ctrl.Pause(ctx); err != nil {
		log.WithError(err).Error("Failed to pause music.")
		return err
	}
	log.Trace("Paused music.")
	return nil
}
