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

// NewSourceService creates a new SourceService.
func NewSourceService(
	src music.Source,
	opts ...basic.Option,
) music.SourceService {
	cfg := basic.BuildOptions(opts...)
	return sourceService{
		src:    src,
		log:    logutil.WithComponent(cfg.Logger, (*sourceService)(nil)),
		tracer: cfg.Tracer,
	}
}

type sourceService struct {
	src    music.Source
	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ music.SourceService = (*sourceService)(nil)

const _defaultLimit = 10

func (svc sourceService) GetTrack(
	ctx context.Context,
	id string,
) (*music.Track, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(sourceService.GetTrack),
	)
	defer span.Finish()

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(sourceService.GetTrack),
		"id":              id,
	}).WithContext(ctx)

	log.Trace("Getting track...")
	t, err := svc.src.GetTrack(ctx, id)
	if err != nil {
		log.WithError(err).Error("Failed to get track.")
		return nil, err
	}
	log.Trace("Got track.")

	return t, nil
}

func (svc sourceService) GetAlbumTracks(
	ctx context.Context,
	id string,
	opts ...music.PaginationOption,
) ([]music.Track, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(sourceService.GetAlbumTracks),
	)
	defer span.Finish()

	opt := music.PaginationOptions{
		Limit:  _defaultLimit,
		Offset: 0,
	}
	for _, apply := range opts {
		apply(&opt)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(sourceService.GetAlbumTracks),
		"id":              id,
		"limit":           opt.Limit,
		"offset":          opt.Offset,
	}).WithContext(ctx)

	log.Trace("Getting album tracks...")
	ts, err := svc.src.GetAlbumTracks(ctx, id, opt)
	if err != nil {
		log.WithError(err).Error("Failed to get album tracks.")
		return nil, err
	}
	log.WithField("tracks", ts).Trace("Got album tracks.")

	return ts, nil
}

func (svc sourceService) GetArtistAlbums(
	ctx context.Context,
	id string,
	opts ...music.PaginationOption,
) ([]music.Album, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(sourceService.GetArtistAlbums),
	)
	defer span.Finish()

	opt := music.PaginationOptions{
		Limit:  _defaultLimit,
		Offset: 0,
	}
	for _, apply := range opts {
		apply(&opt)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(sourceService.GetArtistAlbums),
		"id":              id,
		"limit":           opt.Limit,
		"offset":          opt.Offset,
	}).WithContext(ctx)

	log.Trace("Getting artist albums...")
	as, err := svc.src.GetArtistAlbums(ctx, id, opt)
	if err != nil {
		log.WithError(err).Error("Failed to get artist albums.")
		return nil, err
	}
	log.WithField("albums", as).Trace("Got artist albums.")

	return as, nil
}
