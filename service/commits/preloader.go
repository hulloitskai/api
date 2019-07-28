package commits

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/api/pkg/stream"
	"go.stevenxie.me/api/pkg/zero"
)

type (
	// A Preloader preloads commits for a Service.
	Preloader struct {
		streamer stream.Streamer
		cfg      *PreloaderConfig
		log      logrus.FieldLogger

		mux     sync.Mutex
		commits Commits
		err     error
	}

	// A PreloaderConfig configures a Preloader.
	PreloaderConfig struct {
		Logger logrus.FieldLogger
		Limit  int
	}
)

var _ Service = (*Preloader)(nil)

// NewPreloader creates a new Preloader.
func NewPreloader(
	svc Service,
	interval time.Duration,
	opts ...func(*PreloaderConfig),
) *Preloader {
	cfg := &PreloaderConfig{
		Logger: zero.Logger(),
		Limit:  10,
	}
	for _, opt := range opts {
		opt(cfg)
	}

	cp := &Preloader{
		cfg:     cfg,
		log:     cfg.Logger,
		commits: Commits{},
		streamer: stream.NewPoller(
			func() (zero.Interface, error) {
				return svc.RecentCommits(cfg.Limit)
			},
			interval,
		),
	}
	go cp.run()
	return cp
}

func (cp *Preloader) run() {
	for result := range cp.streamer.Stream() {
		var (
			commits Commits
			err     error
		)

		switch v := result.(type) {
		case error:
			err = v
			cp.log.WithError(err).Error("Failed to load latest commits.")
		case Commits:
			commits = v
		default:
			cp.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("commits: unexpected upstream value '%s'", v)
		}

		cp.mux.Lock()
		cp.commits = commits
		cp.err = err
		cp.mux.Unlock()
	}
}

// Stop stops the Preloader.
func (cp *Preloader) Stop() { cp.streamer.Stop() }

// RecentCommits returns the most recently preloaded commits.
func (cp *Preloader) RecentCommits(limit int) (Commits, error) {
	// Check limit argument.
	if cp.cfg.Limit < limit {
		cp.log.WithFields(logrus.Fields{
			"limit":          cp.cfg.Limit,
			"requestedLimit": limit,
		}).Warn("Commits were requested with a limit greater than the internal " +
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
