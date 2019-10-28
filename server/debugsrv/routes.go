package debugsrv

import (
	"net/http"
	"net/http/pprof"

	"go.stevenxie.me/api/v2/internal"
	"go.stevenxie.me/api/v2/pkg/httputil"
	"go.stevenxie.me/gopkg/name"
)

func (srv Server) registerRoutes() {
	// Register meta route.
	srv.mux.Handle(
		"/",
		httputil.InfoHTTPHandler(
			name.OfTypeFull((*Server)(nil)),
			internal.Version,
		),
	)

	// Register pprof routes.
	{
		const prefix = "/pprof"
		handlers := map[string]http.HandlerFunc{
			"/":        pprof.Index,
			"/cmdline": pprof.Cmdline,
			"/profile": pprof.Profile,
			"/symbol":  pprof.Symbol,
			"/trace":   pprof.Trace,
		}
		for route, handler := range handlers {
			srv.mux.Handle(prefix+route, handler)
		}
	}
}
