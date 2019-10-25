package httpsrv

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/cockroachdb/errors"
	sentry "github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/git"
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/music"
	"go.stevenxie.me/api/productivity"
	"go.stevenxie.me/api/scheduling"
)

// NewServer creates a new Server.
func NewServer(svcs Services, strms Streamers, opts ...ServerOption) *Server {
	cfg := ServerConfig{
		Logger:          logutil.NoopEntry(),
		ComplexityLimit: 5,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Configure logger.
	log := logutil.WithComponent(cfg.Logger, (*Server)(nil))

	// Configure Echo.
	echo := echo.New()
	echo.Logger.SetOutput(ioutil.Discard) // disable logger

	// Configure middleware.
	echo.Pre(middleware.RemoveTrailingSlashWithConfig(
		middleware.TrailingSlashConfig{
			RedirectCode: http.StatusPermanentRedirect,
		},
	))

	// Enable Access-Control-Allow-Origin: * during development.
	if os.Getenv("GOENV") == "development" {
		echo.Use(middleware.CORS())
	}

	// Create and configure server.
	return &Server{
		echo:   echo,
		log:    log,
		sentry: cfg.Sentry,

		svcs:            svcs,
		strms:           strms,
		complexityLimit: cfg.ComplexityLimit,
	}
}

// WithLogger configures a Server to write logs with log.
func WithLogger(log *logrus.Entry) ServerOption {
	return func(cfg *ServerConfig) { cfg.Logger = log }
}

// WithSentry configures a server to capture handler panics with rc.
func WithSentry(c *sentry.Client) ServerOption {
	return func(cfg *ServerConfig) { cfg.Sentry = c }
}

// WithComplexityLimit configures a Server to limit GraphQL queries by
// complexity.
func WithComplexityLimit(limit int) ServerOption {
	return func(cfg *ServerConfig) { cfg.ComplexityLimit = limit }
}

type (
	// Server serves the accounts REST API.
	Server struct {
		echo   *echo.Echo
		log    *logrus.Entry
		sentry *sentry.Client

		svcs  Services
		strms Streamers

		complexityLimit int
	}

	// Services are used to handle server requests.
	Services struct {
		Git          git.Service
		Auth         auth.Service
		About        about.Service
		Music        music.Service
		Transit      transit.Service
		Location     location.Service
		Scheduling   scheduling.Service
		Productivity productivity.Service
	}

	// Streamers are used to handle server streams.
	Streamers struct {
		Music music.Streamer
	}

	// An ServerConfig configures a Server.
	ServerConfig struct {
		Logger *logrus.Entry
		Sentry *sentry.Client

		// Complexity limit for GraphQL queries.
		ComplexityLimit int
	}

	// An ServerOption modifies a ServerConfig.
	ServerOption func(*ServerConfig)
)

// ListenAndServe listens and serves on the specified address.
func (srv *Server) ListenAndServe(addr string) error {
	if addr == "" {
		return errors.New("httpsrv: addr must be non-empty")
	}
	log := srv.log.WithField("addr", addr)

	// Register routes.
	if err := srv.registerRoutes(); err != nil {
		return errors.Wrap(err, "httpsrv: registering routes")
	}

	// Listen for connections.
	log.Info("Listening for connections...")
	return srv.echo.Start(addr)
}

// Shutdown shuts down the server gracefully without interupting any active
// connections.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.echo.Shutdown(ctx)
}
