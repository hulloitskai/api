package locsvc

import (
	"context"
	"time"

	"go.stevenxie.me/api/location/geocode/geoutil"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/location/geocode"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
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
		RegionGeocodeLevel: geocode.CityLevel,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		HistoryService: hist,
		geo:            geo,
		regionLevel:    cfg.RegionGeocodeLevel,
		log:            logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

// WithLogger configures a Service to write logs with log.
func WithLogger(log *logrus.Entry) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.Logger = log }
}

// WithRegionGeocodeLevel configures the geocoding level that a Service uses
// to reverse-geocode my current region.
func WithRegionGeocodeLevel(l geocode.Level) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.RegionGeocodeLevel = l }
}

type (
	service struct {
		location.HistoryService

		geo geocode.Geocoder
		log *logrus.Entry

		regionLevel geocode.Level
	}

	// A ServiceConfig configures a Service.
	ServiceConfig struct {
		Logger             *logrus.Entry
		RegionGeocodeLevel geocode.Level
	}

	// A ServiceOption modifies a ServiceConfig.
	ServiceOption func(*ServiceConfig)
)

var _ location.Service = (*service)(nil)

func (svc service) CurrentPosition(ctx context.Context) (*location.Coordinates, error) {
	log := logutil.
		WithMethod(svc.log, service.CurrentPosition).
		WithContext(ctx)

	segs, err := svc.HistoryService.RecentHistory(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "locsvc: getting recent history")
	}
	log = log.WithField("segments", segs)

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
	log.WithField("position", coords).Trace("Got most recent position.")

	return coords, nil
}

func (svc service) CurrentCity(ctx context.Context) (string, error) {
	log := logutil.
		WithMethod(svc.log, service.CurrentCity).
		WithContext(ctx)

	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		return "", errors.Wrap(err, "locsvc: getting current position")
	}
	log = log.WithField("current_position", coords)
	log.Trace("Got current position.")

	// Reverse-geocode coordinates.
	res, err := svc.geo.ReverseGeocode(
		ctx,
		*coords,
		geocode.WithReverseGeocodeLevel(geocode.CityLevel),
	)
	if err != nil {
		log.WithError(err).Error("Failed to reverse-geocode current position.")
		return "", errors.Wrap(err, "locsvc: reverse-geocoding current position")
	}
	if len(res) == 0 {
		log.Warn("Reverse-geocode query yielded no results.")
		return "", errors.New("locsvc: reverse-geocde search yielded no results")
	}

	// Return city name.
	city := res[0].Place.Address.Label
	log.WithField("city", city).Trace("Reverse-geocoded current city.")

	return city, nil
}

func (svc service) CurrentRegion(
	ctx context.Context,
	opts ...location.CurrentRegionOption,
) (*location.Place, error) {
	var cfg location.CurrentRegionConfig
	for _, opt := range opts {
		opt(&cfg)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey:  name.OfMethod(service.CurrentRegion),
		"include_timezone": cfg.IncludeTimeZone,
	}).WithContext(ctx)

	// Get current position.
	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "locsvc: getting current position")
	}
	log = log.WithField("coordinates", coords)

	// Reverse-geocode region information.
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

	region := &res[0].Place
	log.WithField("region", region).Trace("Reverse-geocoded current region.")

	return region, nil
}

func (svc service) CurrentTimeZone(ctx context.Context) (*time.Location, error) {
	coords, err := svc.CurrentPosition(ctx)
	if err != nil {
		return nil, err
	}
	return geoutil.TimeLocation(ctx, svc.geo, *coords)
}
