package server

import (
	"log"
	"net/http"
	"os"

	"go.uber.org/zap"

	defaults "github.com/mcuadros/go-defaults"
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
func New(v *viper.Viper, logger *zap.SugaredLogger) (*Server, error) {
	// Validate arguments.
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	var err error
	if v == nil {
		if v, err = config.LoadViper(); err != nil {
			return nil, ess.AddCtx("server: loading Viper config", err)
		}
	}

	// Configure self using Viper.
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, ess.AddCtx("server: configuring with Viper", err)
	}
	defaults.SetDefaults(cfg)

	// Configure cron.
	c := cron.New()
	c.ErrorLog = log.New(os.Stderr, "@cron ", log.LstdFlags)

	return &Server{
		Config: cfg,
		Cron:   c,
		viper:  v,
		l:      logger,
	}, nil
}
