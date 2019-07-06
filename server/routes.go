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
		"/productivity",
		handler.ProductivityHandler(srv.productivity, srv.hlog("productivity")),
	)
	e.GET(
		"/availability",
		handler.AvailabilityHandler(srv.availability, srv.hlog("availability")),
	)
	e.GET(
		"/commits",
		handler.RecentCommitsHandler(srv.commits, srv.hlog("recent_commits")),
	)

	// Register location routes.
	location := handler.NewLocationProvider(srv.location)
	e.GET(
		"/location",
		location.RegionHandler(srv.hlog("location")),
	)
	e.GET(
		"/location/history",
		location.RecentHistoryHandler(
			srv.locationAccess,
			srv.hlog("recent_history"),
		),
	)

	// Handle music routes.
	nowplaying := handler.NewNowPlayingProvider(srv.music, srv.music)
	e.GET(
		"/nowplaying",
		nowplaying.RESTHandler(srv.hlog("nowplaying")),
	)
	e.GET(
		"/nowplaying/ws",
		nowplaying.StreamingHandler(srv.hlog("nowplaying_streaming")),
	)

	return nil
}

func (srv *Server) hlog(name string) *logrus.Logger {
	return srv.log.WithField("handler", name).Logger
}
