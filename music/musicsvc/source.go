package musicsvc

import (
	"context"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewSourceService creates a new SourceService.
func NewSourceService(
	src music.Source,
	opts ...svcutil.BasicOption,
) music.SourceService {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return sourceService{
		src: src,
		log: logutil.AddComponent(cfg.Logger, (*sourceService)(nil)),
	}
}

type sourceService struct {
	src music.Source
	log *logrus.Entry
}

var _ music.SourceService = (*sourceService)(nil)

const _defaultLimit = 10

func (svc sourceService) GetTrack(
	ctx context.Context,
	id string,
) (*music.Track, error) {
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
	cfg := music.PaginationConfig{
		Limit:  _defaultLimit,
		Offset: 0,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(sourceService.GetAlbumTracks),
		"id":              id,
		"limit":           cfg.Limit,
		"offset":          cfg.Offset,
	}).WithContext(ctx)

	log.Trace("Getting album tracks...")
	ts, err := svc.src.GetAlbumTracks(ctx, id, cfg)
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
	cfg := music.PaginationConfig{
		Limit:  _defaultLimit,
		Offset: 0,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(sourceService.GetArtistAlbums),
		"id":              id,
		"limit":           cfg.Limit,
		"offset":          cfg.Offset,
	}).WithContext(ctx)

	log.Trace("Getting artist albums...")
	as, err := svc.src.GetArtistAlbums(ctx, id, cfg)
	if err != nil {
		log.WithError(err).Error("Failed to get artist albums.")
		return nil, err
	}
	log.WithField("albums", as).Trace("Got artist albums.")

	return as, nil
}
