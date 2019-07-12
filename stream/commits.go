package stream

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A CommitsPreloader preloads commits while still fulfilling the
	// api.GitCommitService interface.
	CommitsPreloader struct {
		streamer *PollStreamer
		cfg      *CPConfig
		log      *logrus.Logger

		mux     sync.Mutex
		commits []*api.GitCommit
		err     error
	}

	// A CPConfig configures a CommitsPreloader.
	CPConfig struct {
		Logger *logrus.Logger
		Limit  int
	}
)

var _ api.GitCommitsService = (*CommitsPreloader)(nil)

// NewCommitsPreloader creates a new CommitsPreloader.
func NewCommitsPreloader(
	svc api.GitCommitsService,
	interval time.Duration,
	opts ...func(*CPConfig),
) *CommitsPreloader {
	cfg := &CPConfig{
		Logger: zero.Logger(),
		Limit:  10,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	cp := &CommitsPreloader{
		cfg:     cfg,
		log:     cfg.Logger,
		commits: make([]*api.GitCommit, 0),
		streamer: NewPollStreamer(
			func() (zero.Interface, error) {
				return svc.RecentGitCommits(cfg.Limit)
			},
			interval,
		),
	}
	go cp.populateCache()
	return cp
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
			cp.log.WithError(err).Error("Failed to load latest commits.")
		case []*api.GitCommit:
			commits = v
		default:
			cp.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("stream: unexpected upstream value '%s'", v)
		}

		cp.mux.Lock()
		cp.commits = commits
		cp.err = err
		cp.mux.Unlock()
	}
}

// Stop stops the CommitsPreloader.
func (cp *CommitsPreloader) Stop() { cp.streamer.Stop() }

// RecentGitCommits returns the most recently preloaded commits.
func (cp *CommitsPreloader) RecentGitCommits(limit int) ([]*api.GitCommit,
	error) {
	// Check limit argument.
	if cp.cfg.Limit < limit {
		cp.log.WithFields(logrus.Fields{
			"limit":     cp.cfg.Limit,
			"requested": limit,
		}).Warn("Commits were requested with a limit greater than the internal" +
			"limit.")
		limit = cp.cfg.Limit
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
