package gitutil

import (
	"context"
	"time"

	"github.com/stevenxie/api/pkg/zero"

	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
)

// A CommitLoader loads recent Git commits in advance.
type CommitLoader struct {
	// Internal components.
	svc api.GitCommitsService
	ctx context.Context
	log *logrus.Logger

	// Configurable options.
	interval time.Duration
	limit    int

	// Generated values.
	firstload chan struct{}
	commits   []*api.GitCommit
	err       error
}

var _ api.GitCommitsService = (*CommitLoader)(nil)

// NewCommitLoader creates a api.GitCommitsService that preloads commits at
// regular intervals, defaulting to once per minute.
//
// It stops when the provided context is cancelled.
func NewCommitLoader(
	ctx context.Context,
	svc api.GitCommitsService,
	opts ...CLOption,
) *CommitLoader {
	cl := &CommitLoader{
		svc: svc,
		ctx: ctx,
		log: zero.Logger(),

		interval:  time.Minute,
		limit:     10,
		firstload: make(chan struct{}, 1),
	}
	for _, opt := range opts { // evaluate options
		opt(cl)
	}
	go cl.run() // run loading process asynchronously
	return cl
}

// A CLOption configures a CommitLoader.
type CLOption func(*CommitLoader)

// WithInterval configures the amount of time to wait between commit loads.
func WithInterval(interval time.Duration) CLOption {
	return func(cl *CommitLoader) { cl.interval = interval }
}

// WithLimit configures the maximum number of commits to preload.
func WithLimit(limit int) CLOption {
	return func(cl *CommitLoader) { cl.limit = limit }
}

// WithLogger configures the logger that the CommitLoader will write to.
func WithLogger(log *logrus.Logger) CLOption {
	return func(cl *CommitLoader) { cl.log = log }
}

// RecentGitCommits returns the most recently preloaded commits.
func (cl *CommitLoader) RecentGitCommits(limit int) ([]*api.GitCommit, error) {
	// Check limit argument.
	if cl.limit < limit {
		cl.log.WithFields(logrus.Fields{
			"limit":     cl.limit,
			"requested": limit,
		}).Warn("Recent commits were requested with a limit greater than the " +
			"internal limit.")
		limit = cl.limit
	}

	// Guard against requests before first load finishes.
	if cl.commits == nil {
		cl.log.Info("Commits were requested before first load.")
		<-cl.firstload // block until first load completes
	}

	// Since limit is used to slice cl.commits, ensure that:
	//   limit == min(limit, len(cl.commits))
	if limit > len(cl.commits) {
		limit = len(cl.commits)
	}
	return cl.commits[:limit:limit], cl.err
}

func (cl *CommitLoader) run() {
	cl.log.
		WithField("interval", cl.interval.String()).
		Info("Starting commit load loop...")
	trace := time.Now()

	ticker := time.NewTicker(cl.interval)
	cl.loadCommits()
	cl.firstload <- struct{}{} // notify on first load

	cl.log.
		WithField("duration", time.Since(trace).String()).
		Info("Finished loading first set of commits.")

loop:
	for {
		select {
		case <-ticker.C:
			cl.loadCommits()
		case <-cl.ctx.Done():
			ticker.Stop()
			cl.err = cl.ctx.Err()
			break loop
		}
	}
}

func (cl *CommitLoader) loadCommits() {
	cl.log.
		WithField("limit", cl.limit).
		Debug("Loading latest commits...")

	commits, err := cl.svc.RecentGitCommits(cl.limit)
	cl.err = err
	if err != nil {
		cl.log.WithError(err).Error("Failed to load latest commits.")
		return // break early
	}

	cl.commits = commits // save results
}
