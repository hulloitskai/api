package jobserver

import (
	"errors"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/spf13/viper"
)

// A Server handles job requests.
type Server struct {
	*Config
	RedisPool   *redis.Pool
	WorkerPools map[string]*work.WorkerPool
}

// New makes a new Server using cfg.
func New(cfg *Config) *Server {
	if cfg == nil {
		panic(errors.New("work: cannot create Server with nil config"))
	}
	cfg.SetDefaults()

	pool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisAddr)
		},
	}

	return &Server{
		Config:      cfg,
		RedisPool:   pool,
		WorkerPools: make(map[string]*work.WorkerPool),
	}
}

// NewUsing creates a Server, configured using v.
func NewUsing(v *viper.Viper) (*Server, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return New(cfg), nil
}

// Start starts all workers.
func (s *Server) Start() {
	for _, pool := range s.WorkerPools {
		pool.Start()
	}
}

// Stop stops all workers, and closes the underlying Redis connection pool.
func (s *Server) Stop() error {
	for _, pool := range s.WorkerPools {
		pool.Stop()
	}
	return s.RedisPool.Close()
}
