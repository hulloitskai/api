package jobserver

import (
	"github.com/stevenxie/api/internal/util"
	"go.uber.org/zap"
	errors "golang.org/x/xerrors"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/stevenxie/api"
	"github.com/stevenxie/api/processing"
)

// A Server handles background job processing.
type Server struct {
	provider    Provider
	redisPool   *redis.Pool
	workerPools map[string]*work.WorkerPool
	logger      *zap.SugaredLogger

	moodFetcher    *processing.MoodFetcher
	fetchMoodsCron string // default: "0 0-59/5 * * * *"

	started, stopped bool
}

// Provider provides the underlying services required by a Server.
type Provider interface {
	api.MoodSource
	api.MoodService
}

// New creates a new Server.
func New(p Provider, redisAddr string) *Server {
	if redisAddr == "" {
		redisAddr = ":6379"
	}
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", redisAddr)
		},
	}
	return &Server{
		provider:       p,
		redisPool:      redisPool,
		workerPools:    make(map[string]*work.WorkerPool),
		logger:         util.NoopLogger,
		fetchMoodsCron: "0 0-59/5 * * * *",
	}
}

// SetLogger sets the logger used by the Server.
func (srv *Server) SetLogger(logger *zap.SugaredLogger) {
	if logger == nil {
		logger = util.NoopLogger
	}
	srv.logger = logger
	if srv.moodFetcher != nil {
		srv.moodFetcher.SetLogger(logger.Named("moodFetcher"))
	}
}

// SetFetchMoodsCron sets the cron spec for the "fetch moods" job.
func (srv *Server) SetFetchMoodsCron(cron string) { srv.fetchMoodsCron = cron }

// Start starts all workers.
func (srv *Server) Start() error {
	// Validate conditions.
	if srv.stopped {
		return errors.New("jobserver: cannot restart a stopped server")
	}
	if srv.started {
		return nil
	}

	// Register jobs.
	srv.registerJobs()

	// Start workers.
	for _, pool := range srv.workerPools {
		pool.Start()
	}
	srv.started = true
	srv.l().Infof("Job server started.")
	return nil
}

// Stop stops all workers, and closes the underlying Redis connection pool.
func (srv *Server) Stop() error {
	// Validate conditions.
	if srv.stopped {
		return nil
	}

	// Stop workers.
	for _, pool := range srv.workerPools {
		pool.Stop()
	}
	if err := srv.redisPool.Close(); err != nil {
		return errors.Errorf("jobserver: closing redis pool: %w", err)
	}
	srv.stopped = true
	return nil
}

// registerJobs registers jobs to be run by the Server.
func (srv *Server) registerJobs() {
	srv.registerMoodFetcher()
}

func (srv *Server) l() *zap.SugaredLogger { return srv.logger }
