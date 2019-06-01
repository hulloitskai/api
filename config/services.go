package config

import (
	"context"

	"github.com/stevenxie/api/data/github"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/gitutil"
)

// BuildAboutService builds a preconfigured api.AboutService.
func (cfg *Config) BuildAboutService(gr github.GistRepo) *github.AboutService {
	gist := &cfg.About.Gist
	return github.NewAboutService(gr, gist.ID, gist.File)
}

// BuildCommitLoader builds a preconfigured gitutil.CommitLoader.
func (cfg *Config) BuildCommitLoader(
	ctx context.Context,
	svc api.GitCommitsService,
	opts ...gitutil.CLOption,
) *gitutil.CommitLoader {
	var (
		commits = &cfg.Commits
		cfgopts []gitutil.CLOption
	)
	if commits.Limit != nil {
		cfgopts = append(cfgopts, gitutil.WithLimit(*commits.Limit))
	}
	if commits.Interval != nil {
		cfgopts = append(cfgopts, gitutil.WithInterval(*commits.Interval))
	}

	// Build commit loader.
	return gitutil.NewCommitLoader(ctx, svc, append(cfgopts, opts...)...)
}
