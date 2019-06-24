package server

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/stevenxie/api/stream"

	errors "golang.org/x/xerrors"

	"github.com/getsentry/raven-go"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"

	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/zero"
)

// Server serves the accounts REST API.
type Server struct {
	echo *echo.Echo
	log  *logrus.Logger

	about        api.AboutService
	productivity api.ProductivityService
	availability api.AvailabilityService

	commits    *stream.CommitsPreloader
	nowPlaying *stream.NowPlayingStreamer

	// Configurable options.
	nowPlayingPollInterval time.Duration
	commitsPollInterval    time.Duration
	commitsLimit           *int
}

// An Option configures a Server.
type Option func(*Server)

// WithNowPlayingPollInterval sets interval at which the server polls the
// NowPlayingService for updates.
func WithNowPlayingPollInterval(interval time.Duration) Option {
	return func(srv *Server) { srv.nowPlayingPollInterval = interval }
}

// WithGitCommitsPollInterval sets the interval at which the server polls the
// GitCommitsService for updates.
func WithGitCommitsPollInterval(interval time.Duration) Option {
	return func(srv *Server) { srv.commitsPollInterval = interval }
}

// WithGitCommitsLimit sets the maximum number of Git commits to preload.
func WithGitCommitsLimit(limit int) Option {
	return func(srv *Server) { srv.commitsLimit = &limit }
}

// New creates a new Server.
func New(
	about api.AboutService,
	productivity api.ProductivityService,
	availability api.AvailabilityService,
	commits api.GitCommitsService,
	nowPlaying api.NowPlayingService,
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

		nowPlayingPollInterval: 5 * time.Second,
		commitsPollInterval:    time.Minute,
	}
	for _, opt := range opts {
		opt(srv)
	}

	// Build Git commits preloader.
	cpopts := []stream.CPOption{
		stream.WithCPLogger(
			srv.log.WithField("service", "commits_preloader").Logger,
		),
	}
	if srv.commitsLimit != nil {
		cpopts = append(cpopts, stream.WithCPLimit(*srv.commitsLimit))
	}
	srv.commits = stream.NewCommitsPreloader(
		commits,
		srv.commitsPollInterval,
		cpopts...,
	)

	// Build now playing streamer.
	srv.nowPlaying = stream.NewNowPlayingStreamer(
		nowPlaying,
		srv.nowPlayingPollInterval,
		stream.WithNPSLogger(
			srv.log.WithField("service", "nowplaying_streamer").Logger,
		),
	)

	return srv
}

// SetLogger sets the Server's Logger.
func (srv *Server) SetLogger(log *logrus.Logger) { srv.log = log }

// UseRaven configures the Server to capture panic events with the provided
// Raven client.
func (srv *Server) UseRaven(rc *raven.Client) {
	srv.echo.Use(SentryRecoverMiddleware(rc))
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
	// Stop streaming services.
	srv.nowPlaying.Stop()
	srv.commits.Stop()

	// Shut down Echo.
	if err := srv.echo.Shutdown(ctx); err != nil {
		return errors.Errorf("server: shutting down Echo: %w", err)
	}
	return nil
}
