package server

import (
	"net/http"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	errors "golang.org/x/xerrors"
)

// buildHandler builds an http.Handler for srv by wrapping srv.router with
// middleware.
func (srv *Server) buildHandler() http.Handler {
	handle := corsMiddleware(srv.router.ServeHTTP)
	handle = stripTrailingSlash(handle)
	if os.Getenv("GOENV") == "development" {
		handle = loggingMiddleware(handle, srv.l.Desugar())
	}
	return http.HandlerFunc(handle)
}

// corsMiddleware sets the "Access-Control-Allow-Origin" to "*" for all GET and
// HEAD requests.
func corsMiddleware(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "HEAD":
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		handle(w, r)
	}
}

// stripTrailingSlash strips the trailing slash from a request.
func stripTrailingSlash(handle http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") && (len(r.URL.Path) > 1) {
			prefix := r.Header.Get("X-Forwarded-Prefix")
			r.URL.Path = path.Join(prefix, r.URL.Path[:len(r.URL.Path)-1])
			http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
			return
		}
		handle(w, r)
	}
}

func loggingMiddleware(handle http.HandlerFunc,
	logger *zap.Logger) http.HandlerFunc {
	if logger == nil {
		panic(errors.New("server: logger cannot be nil"))
	}

	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug(
			"Incoming request.",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("host", r.Host),
		)
		handle(w, r)
	}
}
