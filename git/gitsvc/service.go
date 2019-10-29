package gitsvc

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/name"

	"go.stevenxie.me/api/v2/git"
	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewService creates a new git.Service.
func NewService(src git.Source, opts ...basic.Option) git.Service {
	cfg := basic.BuildOptions(opts...)
	return service{
		src:    src,
		log:    logutil.WithComponent(cfg.Logger, (*service)(nil)),
		tracer: cfg.Tracer,
	}
}

type service struct {
	src    git.Source
	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ git.Service = (*service)(nil)

func (svc service) RecentCommits(
	ctx context.Context,
	opts ...git.RecentCommitsOption,
) ([]git.Commit, error) {
	span, ctx := opentracing.StartSpanFromContextWithTracer(
		ctx, svc.tracer,
		name.OfFunc(service.RecentCommits),
	)
	defer span.Finish()

	opt := git.RecentCommitsOptions{
		Limit: 10,
	}
	for _, apply := range opts {
		apply(&opt)
	}

	log := svc.log.WithFields(logrus.Fields{
		logutil.MethodKey: name.OfMethod(service.RecentCommits),
		"limit":           opt.Limit,
	}).WithContext(ctx)

	log.Trace("Getting recent commits...")
	cms, err := svc.src.RecentCommits(ctx, opt.Limit)
	if err != nil {
		log.WithError(err).Error("Failed to get recent Git commits.")
		return nil, err
	}
	return cms, nil
}
