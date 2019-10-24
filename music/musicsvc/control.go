package musicsvc

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/basic"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewControlService creates a new music.ControlService.
func NewControlService(
	ctrl music.Controller,
	opts ...basic.Option,
) music.ControlService {
	cfg := basic.BuildConfig(opts...)
	return controlService{
		ctrl:   ctrl,
		log:    logutil.AddComponent(cfg.Logger, (*controlService)(nil)),
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

	var cfg music.PlayConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	if u := cfg.URI; u != nil {
		log.
			WithField("resource", u).
			Trace("Playing the requested resource...")
	} else {
		log.Trace("Resuming the current track...")
	}
	if err := svc.ctrl.Play(ctx, cfg.URI); err != nil {
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
