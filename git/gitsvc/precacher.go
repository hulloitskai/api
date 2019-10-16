package gitsvc

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"

	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/pkg/poll"
)

// NewServicePrecacher creates a new ServicePrecacher.
func NewServicePrecacher(
	svc git.Service,
	interval time.Duration,
	opts ...ServicePrecacherOption,
) ServicePrecacher {
	cfg := ServicePrecacherConfig{
		Logger: logutil.NoopEntry(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	log := logutil.AddComponent(cfg.Logger, (*ServicePrecacher)(nil))
	return ServicePrecacher{
		Service: svc,
		pc: poll.NewPrecacher(
			poll.ProdFunc(func() (zero.Interface, error) {
				return svc.RecentCommits(
					context.Background(),
					func(rcCfg *git.RecentCommitsConfig) {
						if l := cfg.Limit; l != nil {
							rcCfg.Limit = *l
						}
					},
				)
			}),
			interval,
			poll.WithPrecacherLogger(log),
		),
	}
}

type (
	// A ServicePrecacher is a git.Service that precaches recent commits at
	// a regular interval.
	ServicePrecacher struct {
		git.Service
		pc *poll.Precacher
	}

	// A ServicePrecacherConfig configures a ServicePrecacher.
	ServicePrecacherConfig struct {
		Logger *logrus.Entry

		// Limit for the number of recent commits to fetch each time.
		Limit *int
	}

	// A ServicePrecacherOption modifies a ServiceConfig.
	ServicePrecacherOption func(*ServicePrecacherConfig)
)

var _ git.Service = (*ServicePrecacher)(nil)

// RecentCommits implements git.Service for a ServicePrecacher.
func (sp ServicePrecacher) RecentCommits(
	_ context.Context,
	opts ...git.RecentCommitsOption,
) ([]*git.Commit, error) {
	v, err := sp.pc.Results()
	if err != nil {
		return nil, err
	}
	if cms, ok := v.([]*git.Commit); ok {
		cfg := git.RecentCommitsConfig{
			Limit: len(cms),
		}
		for _, opt := range opts {
			opt(&cfg)
		}
		if l := cfg.Limit; len(cms) > l {
			return cms[0:l:l], nil
		}
		return cms, nil
	}
	return nil, nil
}

// Stop stops the ServicePrecacher.
func (sp ServicePrecacher) Stop() { sp.pc.Stop() }
