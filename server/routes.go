package server

import (
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/server/handler"
)

func (srv *Server) registerRoutes() error {
	e := srv.echo

	// Register error handler.
	e.HTTPErrorHandler = handler.ErrorHandler(srv.hlog("error"))

	// Register routes.
	e.GET("/", handler.InfoHandler())
	e.GET("/about", handler.AboutHandler(srv.about, srv.hlog("about")))
	e.GET(
		"/nowplaying",
		handler.NowPlayingHandler(srv.nowPlaying, srv.hlog("nowplaying")),
	)
	e.GET(
		"/productivity",
		handler.ProductivityHandler(srv.productivity, srv.hlog("productivity")),
	)
	e.GET("/commits", handler.RecentCommitsHandler(
		srv.gitCommits,
		srv.hlog("recent_commits"),
	))
	e.GET("/availability", handler.AvailabilityHandler(
		srv.availability,
		srv.hlog("availability"),
	))

	return nil
}

func (srv *Server) hlog(name string) *logrus.Logger {
	return srv.log.WithField("handler", name).Logger
}
