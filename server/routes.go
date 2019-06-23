package server

import (
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/server/handler"
	"github.com/stevenxie/api/stream"
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
	e.GET("/availability", handler.AvailabilityHandler(
		srv.availability,
		srv.hlog("availability"),
	))

	// Handle Git commits (using a streaming preloader).
	cpopts := []stream.CPOption{
		stream.WithCPLogger(
			srv.log.WithField("service", "commits_preloader").Logger,
		),
	}
	if srv.commitsLimit != nil {
		cpopts = append(cpopts, stream.WithCPLimit(*srv.commitsLimit))
	}
	commits := stream.NewCommitsPreloader(
		srv.commits,
		srv.commitsPollInterval,
		cpopts...,
	)
	e.GET("/commits", handler.RecentCommitsHandler(
		commits,
		srv.hlog("recent_commits"),
	))

	// Handle music routes.
	var (
		nps = stream.NewNowPlayingStreamer(
			srv.nowPlaying,
			srv.nowPlayingPollInterval,
			stream.WithNPSLogger(
				srv.log.WithField("service", "nowplaying_streamer").Logger,
			),
		)
		npp = handler.NewNowPlayingProvider(nps, nps)
	)
	e.GET(
		"/nowplaying",
		npp.RESTHandler(srv.hlog("nowplaying_rest")),
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
