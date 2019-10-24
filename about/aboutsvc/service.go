package aboutsvc

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/basic"
)

// NewService creates a new about.Service.
func NewService(
	static about.StaticSource,
	locations location.Service,
	opts ...basic.Option,
) about.Service {
	cfg := basic.BuildConfig(opts...)
	return service{
		static:    static,
		locations: locations,

		log:    logutil.AddComponent(cfg.Logger, (*service)(nil)),
		tracer: cfg.Tracer,
	}
}

type service struct {
	static    about.StaticSource
	locations location.Service

	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ about.Service = (*service)(nil)

func (svc service) GetAbout(ctx context.Context) (*about.About, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.GetAbout),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.GetAbout).
		WithContext(ctx)

	log.Trace("Getting static attributes...")
	static, err := svc.static.GetStatic()
	if err != nil {
		log.WithError(err).Error("Failed to get static attributes.")
		return nil, errors.Wrap(err, "about: getting static attributes")
	}
	log = log.WithField("static_attrs", static)
	log.Trace("Got static attributes.")

	log.Trace("Getting current position...")
	pos, err := svc.locations.CurrentPosition(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current position.")
		return nil, errors.Wrap(err, "about: getting current position")
	}
	log.
		WithField("current_position", pos).
		Trace("Got current position.")

	return &about.About{
		Name:     static.Name,
		Email:    static.Email,
		Type:     static.Type,
		Birthday: static.Birthday,
		Age:      time.Since(static.Birthday),
		IQ:       static.IQ,
		Skills:   static.Skills,
		Location: *pos,
	}, nil
}

func (svc service) GetMasked(ctx context.Context) (*about.Masked, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.GetMasked),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.GetMasked).
		WithContext(ctx)

	log.Trace("Getting static attributes...")
	static, err := svc.static.GetStatic()
	if err != nil {
		log.WithError(err).Error("Failed to get static attributes.")
		return nil, errors.Wrap(err, "about: getting static attributes")
	}
	log = log.WithField("static_attrs", static)
	log.Trace("Got static attributes.")

	log.Trace("Getting current city...")
	city, err := svc.locations.CurrentCity(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current city.")
		return nil, errors.Wrap(err, "about: getting current city")
	}
	log.WithField("city", city).Trace("Got current city.")

	return &about.Masked{
		Name:  static.Name,
		Email: static.Email,
		Type:  static.Type,
		ApproxAge: fmt.Sprintf(
			"about %d years",
			int(time.Since(static.Birthday).Hours())/(365*24),
		),
		IQ:          static.IQ,
		Skills:      static.Skills,
		Whereabouts: city,
	}, nil
}
