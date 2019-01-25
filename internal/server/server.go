package server

import (
	"context"
	"net/http"

	"go.uber.org/zap"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/config"
	"github.com/stevenxie/api/internal/server/routes"
	"github.com/stevenxie/api/pkg/data/postgres"
	ess "github.com/unixpickle/essentials"
)

// Server serves a REST API for interacting with personal data.
type Server struct {
	WebServer *http.Server
	DB        *gorm.DB
	*Config

	viper *viper.Viper
	l     *zap.SugaredLogger
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
	if err = cfg.SetDefaults(); err != nil {
		return nil, ess.AddCtx("server: setting config defaults", err)
	}

	return &Server{
		Config: cfg,
		viper:  viper,
		l:      logger,
	}, nil
}

// ListenAndServe starts the server on the specified address.
func (s *Server) ListenAndServe(addr string) error {
	// Configure DB.
	db, err := postgres.OpenUsing(s.viper)
	if err != nil {
		return ess.AddCtx("server: opening DB connection", err)
	}
	s.DB = db

	// Make and configure router.
	cfg := routes.Config{
		Logger: s.l.Named("routes"),
		DB:     db,
	}
	router, err := routes.NewRouter(&cfg)
	if err != nil {
		return ess.AddCtx("server: creating router", err)
	}

	s.WebServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return s.WebServer.ListenAndServe()
}

// Shutdown gracefully shuts down the Server, closing all existing connections.
func (s *Server) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(),
		s.ShutdownTimeout)
	defer cancel()

	// Shutdown webserver.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := s.WebServer.Shutdown(ctx); err != nil {
			return ess.AddCtx("server: shutting down internal webserver", err)
		}
	}

	// Close DB connection.
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		err := s.DB.Close()
		return ess.AddCtx("server: shutting down DB", err)
	}
}
