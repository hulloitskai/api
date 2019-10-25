package locsvc

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/api/location/geocode/geoutil"
)

// NewService creates a new location.Service using a HistoryService and
// a geocode.Geocoder.
func NewService(
	hist location.HistoryService,
	geo geocode.Geocoder,
	opts ...ServiceOption,
) location.Service {
	cfg := ServiceConfig{
		Logger:             logutil.NoopEntry(),
		Tracer:             new(opentracing.NoopTracer),
		RegionGeocodeLevel: geocode.CityLevel,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		HistoryService: hist,
		geo:            geo,
		regionLevel:    cfg.RegionGeocodeLevel,

		log:    logutil.WithComponent(cfg.Logger, (*service)(nil)),
		tracer: cfg.Tracer,
	}
}

// WithLogger configures a Service to write logs with log.
func WithLogger(log *logrus.Entry) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.Logger = log }
}

// WithTracer configures a Service to trace calls with t.
func WithTracer(t opentracing.Tracer) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.Tracer = t }
}

// WithRegionGeocodeLevel configures the geocoding level that a Service uses
// to reverse-geocode my current region.
func WithRegionGeocodeLevel(l geocode.Level) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.RegionGeocodeLevel = l }
}

type (
	service struct {
		location.HistoryService
		geo         geocode.Geocoder
		regionLevel geocode.Level

		log    *logrus.Entry
		tracer opentracing.Tracer
	}

	// A ServiceConfig configures a Service.
	ServiceConfig struct {
		Logger *logrus.Entry
		Tracer opentracing.Tracer

		RegionGeocodeLevel geocode.Level
	}

	// A ServiceOption modifies a ServiceConfig.
	ServiceOption func(*ServiceConfig)
)

var _ location.Service = (*service)(nil)

func (svc service) CurrentPosition(ctx context.Context) (*location.Coordinates, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.CurrentPosition),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.CurrentPosition).
		WithContext(ctx)

	log.Trace("Getting recent location history...")
	segs, err := svc.HistoryService.RecentHistory(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "locsvc: getting recent history")
	}
	log = log.WithField("segments", segs)
	log.Trace("Got location history segments.")

	if len(segs) == 0 {
		log.Error("No history segments found.")
		return nil, errors.New("locsvc: no history segments found")
	}
	seg := segs[len(segs)-1]

	coords := latestCoordinates(&seg)
	if coords == nil {
		log.Error("No coordinates in latest history segment0w.")
		return nil, errors.New("locsvc: no coordinates in latest history segment")
	}
	return coords, nil
}

func (svc service) CurrentCity(ctx context.Context) (string, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.CurrentCity),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.CurrentCity).
		WithContext(ctx)

	log.Trace("Getting current position...")
	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		return "", errors.Wrap(err, "locsvc: getting current position")
	}
	log = log.WithField("current_position", coords)
	log.Trace("Got current position.")

	// Reverse-geocode coordinates.
	log.Trace("Reverse-geocoding coordinates.")
	res, err := svc.geo.ReverseGeocode(
		ctx,
		*coords,
		geocode.ReverseWithLevel(geocode.CityLevel),
	)
	if err != nil {
		log.WithError(err).Error("Failed to reverse-geocode current position.")
		return "", errors.Wrap(err, "locsvc: reverse-geocoding current position")
	}
	if len(res) == 0 {
		log.Warn("Reverse-geocode query yielded no results.")
		return "", errors.New("locsvc: reverse-geocde search yielded no results")
	}
	log.WithField("result", res).Trace("Got result from geocoder.")

	// Return city name.
	return res[0].Place.Address.Label, nil
}

func (svc service) CurrentRegion(
	ctx context.Context,
	opts ...location.CurrentRegionOption,
) (*location.Place, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.CurrentRegion),
	)

	defer span.Finish()
	var cfg location.CurrentRegionConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey:  name.OfMethod(service.CurrentRegion),
		"include_timezone": cfg.IncludeTimeZone,
	}).WithContext(ctx)

	// Get current position.
	log.Trace("Getting current position.")
	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "locsvc: getting current position")
	}
	log = log.WithField("position", coords)
	log.Trace("Got current position.")

	// Reverse-geocode region information.
	log.Trace("Reverse-geocoding coordinates.")
	res, err := svc.geo.ReverseGeocode(
		ctx,
		*coords,
		func(rgCfg *geocode.ReverseGeocodeConfig) {
			rgCfg.Level = svc.regionLevel
			rgCfg.IncludeShape = true
			rgCfg.IncludeTimeZone = cfg.IncludeTimeZone
		},
	)
	if err != nil {
		log.WithError(err).Error("Failed to reverse-geocode current position.")
		return nil, errors.Wrap(err, "locsvc: reverse-geocoding current position")
	}
	if len(res) == 0 {
		log.Warn("Reverse-geocode query yielded no results.")
		return nil, errors.New("locsvc: reverse-geocde search yielded no results")
	}
	log.WithField("results", res).Trace("Got results from geocoder.")
	return &res[0].Place, nil
}

func (svc service) CurrentTimeZone(ctx context.Context) (*time.Location, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.CurrentTimeZone),
	)
	defer span.Finish()

	log := logutil.
		WithMethod(svc.log, service.CurrentTimeZone).
		WithContext(ctx)

	log.Trace("Getting current position...")
	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		log.WithError(err).Error("Failed to get current position.")
		return nil, err
	}
	log.WithField("position", coords).Trace("Got current position.")

	return geoutil.TimeLocation(ctx, svc.geo, *coords)
}
