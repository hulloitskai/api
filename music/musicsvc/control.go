package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewControlService creates a new music.ControlService.
func NewControlService(
	ctrl music.Controller,
	opts ...basic.Option,
) music.ControlService {
	opt := basic.BuildOptions(opts...)
	return controlService{
		ctrl: ctrl,
		log:  logutil.WithComponent(opt.Logger, (*controlService)(nil)),
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
	log := logutil.
		WithMethod(svc.log, controlService.Play).
		WithContext(ctx)

	var opt music.PlayOptions
	for _, apply := range opts {
		apply(&opt)
	}

	s := opt.Selector
	if s != nil {
		log.
			WithField("selector", s).
			Trace("Playing the selected resource...")
	} else {
		log.Trace("Resuming the current track...")
	}
	if err := svc.ctrl.Play(ctx, s); err != nil {
		log.WithError(err).Error("Failed to play resource.")
		return err
	}
	return nil
}

func (svc controlService) Pause(ctx context.Context) error {
	log := logutil.
		WithMethod(svc.log, controlService.Pause).
		WithContext(ctx)

	log.Trace("Pausing the current track...")
	if err := svc.ctrl.Pause(ctx); err != nil {
		log.WithError(err).Error("Failed to pause music.")
		return err
	}
	return nil
}
