package server

import (
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"

	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/config"
	"github.com/stevenxie/api/internal/data"
	ess "github.com/unixpickle/essentials"
)

// Server serves a REST API for interacting with personal data.
type Server struct {
	*Config

	WebServer *http.Server
	Repos     *data.RepoSet
	Cron      *cron.Cron

	drivers *data.DriverSet
	viper   *viper.Viper
	l       *zap.SugaredLogger
}

// New returnsa new Server.
func New(logger *zap.SugaredLogger) (*Server, error) {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	// Load Viper config.
	viper, err := config.LoadViper()
	if err != nil {
		return nil, ess.AddCtx("server: loading Viper config", err)
	}

	// Configure self using Viper.
	cfg, err := ConfigFromViper(viper)
	if err != nil {
		return nil, ess.AddCtx("server: configuring with Viper", err)
	}
	cfg.SetDefaults()

	// Configure cron.
	c := cron.New()
	c.ErrorLog = log.New(os.Stderr, "@cron ", log.LstdFlags)

	return &Server{
		Config: cfg,
		Cron:   c,
		viper:  viper,
		l:      logger,
	}, nil
}
