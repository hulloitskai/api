package musicsvc

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/music"
	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewControlService creates a new music.ControlService.
func NewControlService(
	ctrl music.Controller,
	opts ...basic.Option,
) music.ControlService {
	cfg := basic.BuildOptions(opts...)
	return controlService{
		ctrl:   ctrl,
		log:    logutil.WithComponent(cfg.Logger, (*controlService)(nil)),
		tracer: cfg.Tracer,
	}
}

type controlService struct {
	ctrl   music.Controller
	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ music.ControlService = (*controlService)(nil)

func (svc controlService) Play(
	ctx context.Context,
	opts ...music.PlayOption,
) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(controlService.Play),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, controlService.Play).
		WithContext(ctx)

	var opt music.PlayOptions
	for _, apply := range opts {
		apply(&opt)
	}

	if u := opt.URI; u != nil {
		log.
			WithField("resource", u).
			Trace("Playing the requested resource...")
	} else {
		log.Trace("Resuming the current track...")
	}
	if err := svc.ctrl.Play(ctx, opt.URI); err != nil {
		log.WithError(err).Error("Failed to play resource.")
		return err
	}
	return nil
}

func (svc controlService) Pause(ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(controlService.Pause),
	)
	defer span.Finish()

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
