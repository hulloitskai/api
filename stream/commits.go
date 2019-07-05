package stream

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A CommitsPreloader preloads commits while still fulfilling the
	// api.GitCommitService interface.
	CommitsPreloader struct {
		streamer *PollStreamer
		log      *logrus.Logger

		// Configurable options.
		limit int

		mux     sync.Mutex
		commits []*api.GitCommit
		err     error
	}

	// A CPOption configures a CommitsPreloader.
	CPOption func(*CommitsPreloader)
)

var _ api.GitCommitsService = (*CommitsPreloader)(nil)

// NewCommitsPreloader creates a new CommitsPreloader.
func NewCommitsPreloader(
	svc api.GitCommitsService,
	interval time.Duration,
	opts ...CPOption,
) *CommitsPreloader {
	cp := &CommitsPreloader{
		log:     zero.Logger(),
		limit:   10,
		commits: make([]*api.GitCommit, 0),
	}
	for _, opt := range opts {
		opt(cp)
	}

	// Configure streamer.
	action := func() (zero.Interface, error) {
		return svc.RecentGitCommits(cp.limit)
	}
	cp.streamer = NewPollStreamer(action, interval)

	go cp.populateCache()
	return cp
}

// WithCPLogger configures a CommitPreloader's logger.
func WithCPLogger(log *logrus.Logger) CPOption {
	return func(cp *CommitsPreloader) { cp.log = log }
}

// WithCPLimit sets the maximum number of commits that a CommitPreloader will
// preload.
func WithCPLimit(limit int) CPOption {
	return func(cp *CommitsPreloader) { cp.limit = limit }
}

func (cp *CommitsPreloader) populateCache() {
	for result := range cp.streamer.Stream() {
		var (
			commits []*api.GitCommit
			err     error
		)

		switch v := result.(type) {
		case error:
			err = v
		case []*api.GitCommit:
			commits = v
		}

		cp.mux.Lock()
		cp.commits = commits
		cp.err = err
		cp.mux.Unlock()
	}
}

// RecentGitCommits returns the most recently preloaded commits.
func (cp *CommitsPreloader) RecentGitCommits(limit int) ([]*api.GitCommit,
	error) {
	// Check limit argument.
	if cp.limit < limit {
		cp.log.WithFields(logrus.Fields{
			"limit":     cp.limit,
			"requested": limit,
		}).Warn("Commits were requested with a limit greater than the internal" +
			"limit.")
		limit = cp.limit
	}

	// Guard access to cp.commits and cp.err.
	cp.mux.Lock()
	defer cp.mux.Unlock()

	// Since limit is used to slice cl.commits, ensure that:
	//   limit == min(limit, len(cl.commits))
	if limit > len(cp.commits) {
		limit = len(cp.commits)
	}
	return cp.commits[:limit:limit], cp.err
}

// Stop stops the CommitsPreloader.
func (cp *CommitsPreloader) Stop() { cp.streamer.Stop() }
