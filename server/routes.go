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
	e.GET("/about", handler.AboutHandler(srv.hlogger("about"), srv.info))
	e.GET(
		"/nowplaying",
		handler.NowPlayingHandler(srv.hlogger("nowplaying"), srv.currentlyPlaying),
	)
	e.GET(
		"/productivity",
		handler.ProductivityHandler(srv.hlogger("productivity"), srv.productivity),
	)
	e.GET("/commits", handler.RecentCommitsHandler(
		srv.hlogger("recent_commits"),
		srv.recentCommits,
	))

	return nil
}

func (srv *Server) hlogger(name string) zerolog.Logger {
	return srv.logger.With().Str("handler", name).Logger()
}
