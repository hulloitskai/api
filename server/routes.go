package server

import (
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/server/handler"
)

func (srv *Server) registerRoutes() error {
	e := srv.echo

	// Register error handler.
	e.HTTPErrorHandler = handler.ErrorHandler(srv.hlogger("error"))

	// Register routes.
	e.GET("/", handler.InfoHandler())
	e.GET("/about", handler.AboutHandler(srv.about, srv.hlogger("about")))
	e.GET(
		"/nowplaying",
		handler.NowPlayingHandler(srv.nowPlaying, srv.hlogger("nowplaying")),
	)
	e.GET(
		"/productivity",
		handler.ProductivityHandler(srv.productivity, srv.hlogger("productivity")),
	)
	e.GET("/commits", handler.RecentCommitsHandler(
		srv.gitCommits,
		srv.hlogger("recent_commits"),
	))
	e.GET("/availability", handler.AvailabilityHandler(
		srv.availability,
		srv.hlogger("availability"),
	))

	return nil
}

func (srv *Server) hlogger(name string) zerolog.Logger {
	return srv.logger.With().Str("handler", name).Logger()
}
