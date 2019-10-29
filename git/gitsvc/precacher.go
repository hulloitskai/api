package gitsvc

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
	"go.stevenxie.me/gopkg/zero"

	"go.stevenxie.me/api/v2/git"
	"go.stevenxie.me/api/v2/pkg/poll"
)

// NewServicePrecacher creates a new ServicePrecacher.
func NewServicePrecacher(
	svc git.Service,
	interval time.Duration,
	opts ...ServicePrecacherOption,
) ServicePrecacher {
	opt := ServicePrecacherOptions{
		Logger: logutil.NoopEntry(),
	}
	for _, apply := range opts {
		apply(&opt)
	}
	log := logutil.WithComponent(opt.Logger, (*ServicePrecacher)(nil))
	return ServicePrecacher{
		Service: svc,
		pc: poll.NewPrecacher(
			poll.ProdFunc(func() (zero.Interface, error) {
				return svc.RecentCommits(
					context.Background(),
					func(rcOpt *git.RecentCommitsOptions) {
						if l := opt.Limit; l != nil {
							rcOpt.Limit = *l
						}
					},
				)
			}),
			interval,
			poll.PrecacherWithLogger(log),
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

	// A ServicePrecacherOptions configures a ServicePrecacher.
	ServicePrecacherOptions struct {
		Logger *logrus.Entry

		// Limit for the number of recent commits to fetch each time.
		Limit *int
	}

	// A ServicePrecacherOption modifies a ServicePrecacherOptions.
	ServicePrecacherOption func(*ServicePrecacherOptions)
)

var _ git.Service = (*ServicePrecacher)(nil)

// RecentCommits implements git.Service for a ServicePrecacher.
func (sp ServicePrecacher) RecentCommits(
	_ context.Context,
	opts ...git.RecentCommitsOption,
) ([]git.Commit, error) {
	v, err := sp.pc.Results()
	if err != nil {
		return nil, err
	}
	if cms, ok := v.([]git.Commit); ok {
		opt := git.RecentCommitsOptions{
			Limit: len(cms),
		}
		for _, apply := range opts {
			apply(&opt)
		}
		if l := opt.Limit; len(cms) > l {
			return cms[0:l:l], nil
		}
		return cms, nil
	}
	return nil, nil
}

// Stop stops the ServicePrecacher.
func (sp ServicePrecacher) Stop() { sp.pc.Stop() }
