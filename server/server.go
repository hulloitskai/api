package server

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/getsentry/raven-go"
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
		music        api.MusicService

		location       api.LocationService
		locationAccess api.LocationAccessService
	}

	// An Config configures a Server.
	Config struct {
		Logger *logrus.Logger
		Raven  *raven.Client
	}
)

// New creates a new Server.
func New(
	about api.AboutService,
	availability api.AvailabilityService,
	commits api.GitCommitsService,
	music api.MusicService,
	productivity api.ProductivityService,

	location api.LocationService,
	locationAccess api.LocationAccessService,

	opts ...func(*Config),
) *Server {
	cfg := Config{Logger: zero.Logger()}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Configure echo.
	echo := echo.New()
	echo.Logger.SetOutput(ioutil.Discard) // disable logger

	// Configure middleware.
	echo.Pre(middleware.RemoveTrailingSlashWithConfig(
		middleware.TrailingSlashConfig{
			RedirectCode: http.StatusPermanentRedirect,
		},
	))
	if cfg.Raven != nil {
		echo.Use(SentryRecoverMiddleware(cfg.Raven, cfg.Logger))
	}

	// Enable Access-Control-Allow-Origin: * during development.
	if os.Getenv("GOENV") == "development" {
		echo.Use(middleware.CORS())
	}

	// Create and configure server.
	return &Server{
		echo: echo,
		log:  cfg.Logger,

		about:        about,
		availability: availability,
		commits:      commits,
		music:        music,
		productivity: productivity,

		location:       location,
		locationAccess: locationAccess,
	}
}

// ListenAndServe listens and serves on the specified address.
func (srv *Server) ListenAndServe(addr string) error {
	if addr == "" {
		return errors.New("server: addr must be non-empty")
	}

	// Register routes.
	if err := srv.registerRoutes(); err != nil {
		return errors.Wrap(err, "server: registering routes")
	}

	// Listen for connections.
	srv.log.WithField("addr", addr).Info("Listening for connections...")
	return srv.echo.Start(addr)
}

// Shutdown shuts down the server gracefully without interupting any active
// connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	if err := srv.echo.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "server: shutting down Echo")
	}
	return nil
}
