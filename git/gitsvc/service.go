package gitsvc

import (
	"context"

	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/pkg/svcutil"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"
)

// NewService creates a new git.Service.
func NewService(src git.Source, opts ...svcutil.BasicOption) git.Service {
	cfg := svcutil.BasicConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return service{
		src: src,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

type service struct {
	src git.Source
	log *logrus.Entry
}

var _ git.Service = (*service)(nil)

func (svc service) RecentCommits(
	ctx context.Context,
	opts ...git.RecentCommitsOption,
) ([]*git.Commit, error) {
	cfg := git.RecentCommitsConfig{
		Limit: 10,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.RecentCommits),
		"limit":           cfg.Limit,
	}).WithContext(ctx)

	cms, err := svc.src.RecentCommits(ctx, cfg.Limit)
	if err != nil {
		log.WithError(err).Error("Failed to get recent git.")
		return nil, err
	}
	log.WithField("commits", cms).Trace("Got recent git.")

	return cms, nil
}
