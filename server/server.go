package server

import (
	"net/http"

	"go.uber.org/zap"

	defaults "github.com/mcuadros/go-defaults"
	"github.com/stevenxie/api"
	"github.com/stevenxie/api/server/routes"
	ess "github.com/unixpickle/essentials"
)

// Server serves a REST API for interacting with personal data.
type Server struct {
	*Config
	Provider  api.ServiceProvider
	WebServer *http.Server

	l *zap.SugaredLogger
}

// New returnsa new Server.
func New(provider api.ServiceProvider, logger *zap.SugaredLogger,
	cfg *Config) (*Server, error) {
	// Validate arguments.
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	if cfg == nil {
		cfg = new(Config)
	}
	defaults.SetDefaults(cfg)

	// Build router.
	router, err := routes.NewRouter(provider, logger.Named("router"))
	if err != nil {
		return nil, ess.AddCtx("server: creating router", err)
	}

	return &Server{
		Config:    cfg,
		Provider:  provider,
		WebServer: &http.Server{Handler: router},
		l:         logger,
	}, nil
}
