package routes

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api"
	"go.uber.org/zap"
)

// Router matches requests to the routes defined in this package.
type Router struct {
	Provider api.ServiceProvider

	hr httprouter.Router
	l  *zap.SugaredLogger
}

// NewRouter returns a new Router.
func NewRouter(provider api.ServiceProvider,
	logger *zap.SugaredLogger) (*Router, error) {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	hr := httprouter.New()
	r := &Router{
		Provider: provider,
		hr:       *hr,
		l:        logger,
	}
	r.registerRoutes()
	return r, nil
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handle := corsMiddleware(r.hr.ServeHTTP)
	handle(w, req)
}

func (r *Router) registerRoutes() {
	router := &r.hr
	registerIndex(router, r.l.Named("index"))
	registerMoods(router, r.Provider, r.l.Named("moods"))
}
