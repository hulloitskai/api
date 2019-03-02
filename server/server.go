package server

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api"
	"github.com/stevenxie/api/internal/util"
)

// Server serves a REST API for the services defined in package api.
type Server struct {
	provider Provider
	httpsrv  *http.Server
	router   *httprouter.Router
	logger   *zap.SugaredLogger

	shutdownTimeout time.Duration
}

// Provider provides the underlying services required by a Router.
type Provider interface {
	api.MoodService
}

// New creates a new Server.
func New(p Provider) *Server {
	router := httprouter.New()
	router.RedirectTrailingSlash = false

	srv := &Server{
		provider: p,
		router:   router,
		httpsrv:  new(http.Server),
		logger:   util.NoopLogger,
	}
	srv.registerRoutes()
	return srv
}

// SetShutdownTimeout sets the shutdown timeout for srv.
func (srv *Server) SetShutdownTimeout(timeout time.Duration) {
	srv.shutdownTimeout = timeout
}

// SetLogger sets a zap.SugaredLogger for srv.
func (srv *Server) SetLogger(logger *zap.SugaredLogger) {
	if logger == nil {
		logger = util.NoopLogger
	}
	srv.logger = logger
}

// ListenAndServe starts the server, and listens for connections on addr.
func (srv *Server) ListenAndServe(addr string) error {
	srv.httpsrv.Handler = srv.buildHandler()
	srv.httpsrv.Addr = addr
	srv.l().Infof("Listening on address '%s'...", addr)
	return srv.httpsrv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (srv *Server) Shutdown() error {
	ctx, cancel := srv.shutdownContext()
	defer cancel()
	return srv.httpsrv.Shutdown(ctx)
}

// shutdownContext creates a context.Context used for shutting down a server.
func (srv *Server) shutdownContext() (context.Context, context.CancelFunc) {
	return util.ContextWithTimeout(srv.shutdownTimeout)
}

func (srv *Server) l() *zap.SugaredLogger { return srv.logger }
