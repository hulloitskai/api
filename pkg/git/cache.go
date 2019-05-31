package git

import (
	"context"
	"time"

	"github.com/rs/zerolog"
)

const (
	updatingRecentCommitsInterval = 1 * time.Minute
	updatingRecentCommitsLimit    = 10
)

type updatingRecentCommitsCache struct {
	svc    RecentCommitsService
	ctx    context.Context
	logger zerolog.Logger

	commits []*Commit
	err     error
}

// WithUpdatingCache creates a RecentCommitsService that caches recent commits,
// and updates its internal cache every minute.
//
// It stops when the provided context is cancelled.
func WithUpdatingCache(
	ctx context.Context,
	rcs RecentCommitsService,
	l zerolog.Logger,
) RecentCommitsService {
	cache := &updatingRecentCommitsCache{
		svc:    rcs,
		ctx:    ctx,
		logger: l,
	}
	go cache.run()
	return cache
}

// RecentCommits returns the cached slice of recent commits.
func (c *updatingRecentCommitsCache) RecentCommits(limit int) (
	[]*Commit, error) {
	if c.commits == nil {
		return c.svc.RecentCommits(limit)
	}

	if limit > updatingRecentCommitsLimit {
		limit = updatingRecentCommitsLimit
	}
	if limit > len(c.commits) {
		limit = len(c.commits)
	}
	return c.commits[:limit:limit], c.err
}

func (c *updatingRecentCommitsCache) run() {
	ticker := time.NewTicker(updatingRecentCommitsInterval)
	c.l().Info().Msg("Starting cache update loop...")
	c.update()

loop:
	for {
		select {
		case <-ticker.C:
			c.update()
		case <-c.ctx.Done():
			ticker.Stop()
			c.err = c.ctx.Err()
			break loop
		}
	}
}

func (c *updatingRecentCommitsCache) l() *zerolog.Logger { return &c.logger }

func (c *updatingRecentCommitsCache) update() {
	c.l().Debug().Msg("Updating cache with latest commits...")
	commits, err := c.svc.RecentCommits(updatingRecentCommitsLimit)
	if err != nil {
		c.l().Err(err).Msg("Failed to update cache.")
		c.err = err
		return
	}
	c.commits = commits
}
