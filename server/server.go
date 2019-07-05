package server

import (
	"context"
	"io/ioutil"
	"os"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// Server serves the accounts REST API.
	Server struct {
		echo *echo.Echo
		log  *logrus.Logger

		about        api.AboutService
		productivity api.ProductivityService
		availability api.AvailabilityService
		commits      api.GitCommitsService
		nowPlaying   api.NowPlayingStreamingService
	}

	// An Option configures a Server.
	Option func(*Server)
)

// New creates a new Server.
func New(
	about api.AboutService,
	productivity api.ProductivityService,
	availability api.AvailabilityService,
	nowPlaying api.NowPlayingStreamingService,
	commits api.GitCommitsService,
	opts ...Option,
) *Server {
	// Configure echo.
	echo := echo.New()
	echo.Logger.SetOutput(ioutil.Discard) // disable logger
	echo.Use(middleware.Recover())

	// Enable Access-Control-Allow-Origin: * during development.
	if os.Getenv("GOENV") == "development" {
		echo.Use(middleware.CORS())
	}

	// Create and configure server.
	srv := &Server{
		echo: echo,
		log:  zero.Logger(),

		about:        about,
		productivity: productivity,
		availability: availability,
		nowPlaying:   nowPlaying,
		commits:      commits,
	}
	for _, opt := range opts {
		opt(srv)
	}
	return srv
}

// ListenAndServe listens and serves on the specified address.
func (srv *Server) ListenAndServe(addr string) error {
	if addr == "" {
		return errors.New("server: addr must be non-empty")
	}

	// Register routes.
	if err := srv.registerRoutes(); err != nil {
		return errors.Errorf("server: registering routes: %w", err)
	}

	// Listen for connections.
	srv.log.WithField("addr", addr).Info("Listening for connections...")
	return srv.echo.Start(addr)
}

// Shutdown shuts down the server gracefully without interupting any active
// connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	if err := srv.echo.Shutdown(ctx); err != nil {
		return errors.Errorf("server: shutting down Echo: %w", err)
	}
	return nil
}
