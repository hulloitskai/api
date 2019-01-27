package routes

import (
	"errors"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/stevenxie/api/internal/data"
	"go.uber.org/zap"
)

// Router matches requests to the routes defined in this package.
type Router struct {
	Repos *data.RepoSet
	hr    httprouter.Router
	l     *zap.SugaredLogger
}

// NewRouter returns a new Router.
func NewRouter(repos *data.RepoSet, logger *zap.SugaredLogger) (*Router, error) {
	if repos == nil {
		panic(errors.New("server: cannot make router with nil data.RepoSet"))
	}
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}

	hr := httprouter.New()
	r := &Router{
		Repos: repos,
		hr:    *hr,
		l:     logger,
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
	registerMoods(router, r.Repos.MoodRepo, r.l.Named("moods"))
}
