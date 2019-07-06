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
	e.GET("/productivity", handler.ProductivityHandler(
		srv.productivity,
		srv.hlog("productivity")),
	)
	e.GET("/availability", handler.AvailabilityHandler(
		srv.availability,
		srv.hlog("availability"),
	))
	e.GET("/commits", handler.RecentCommitsHandler(
		srv.commits,
		srv.hlog("recent_commits"),
	))
	e.GET("/location", handler.LocationHandler(
		srv.location,
		srv.hlog("location"),
	))

	// Handle music routes.
	npp := handler.NewNowPlayingProvider(srv.music, srv.music)
	e.GET(
		"/nowplaying",
		npp.RESTHandler(srv.hlog("nowplaying")),
	)
	e.GET(
		"/nowplaying/ws",
		npp.StreamingHandler(srv.hlog("nowplaying_streaming")),
	)

	return nil
}

func (srv *Server) hlog(name string) *logrus.Logger {
	return srv.log.WithField("handler", name).Logger
}
