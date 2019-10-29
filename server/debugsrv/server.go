package debugsrv

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"

	"go.stevenxie.me/api/v2/pkg/basic"
)

// NewServer creates a new Server.
func NewServer(opts ...basic.Option) Server {
	cfg := basic.BuildOptions(opts...)
	mux := http.NewServeMux()
	return Server{
		http: &http.Server{Handler: mux},
		mux:  mux,
		log:  logutil.WithComponent(cfg.Logger, (*Server)(nil)),
	}
}

// A Server handles debugging requests.
type Server struct {
	http *http.Server
	mux  *http.ServeMux
	log  *logrus.Entry
}

// ListenAndServe handles incoming HTTP requests to addr.
func (srv Server) ListenAndServe(addr string) error {
	if addr == "" {
		return errors.New("debugsrv: addr must be non-empty")
	}
	srv.http.Addr = addr
	log := srv.log.WithField("addr", addr)

	// Register routes.
	srv.registerRoutes()

	// Listen for connections.
	log.Info("Listening for connections...")
	return srv.http.ListenAndServe()
}

// Shutdown forwards a definition.
func (srv Server) Shutdown(ctx context.Context) error {
	return srv.http.Shutdown(ctx)
}
