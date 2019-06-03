package server

import (
	"context"
	"io/ioutil"
	"os"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/api"
)

// Server serves the accounts REST API.
type Server struct {
	echo   *echo.Echo
	logger zerolog.Logger

	about        api.AboutService
	productivity api.ProductivityService
	availability api.AvailabilityService
	gitCommits   api.GitCommitsService
	nowPlaying   api.NowPlayingService
}

// New creates a new Server.
func New(
	about api.AboutService,
	productivity api.ProductivityService,
	availability api.AvailabilityService,
	gitCommits api.GitCommitsService,
	nowPlaying api.NowPlayingService,
	l zerolog.Logger,
) *Server {
	// Configure echo.
	echo := echo.New()
	echo.Logger.SetOutput(ioutil.Discard) // disable logger

	// Enable Access-Control-Allow-Origin: * during development.
	if os.Getenv("GOENV") == "development" {
		echo.Use(middleware.CORS())
	}

	return &Server{
		echo:         echo,
		logger:       l,
		about:        about,
		productivity: productivity,
		availability: availability,
		gitCommits:   gitCommits,
		nowPlaying:   nowPlaying,
	}
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
	srv.logger.Info().Str("addr", addr).Msg("Listening for connections...")
	return srv.echo.Start(addr)
}

// Shutdown shuts down the server gracefully without interupting any active
// connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.echo.Shutdown(ctx)
}
