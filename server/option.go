package server

import (
	"time"

	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// WithNowPlayingPollInterval sets interval at which the server polls the
// NowPlayingService for updates.
func WithNowPlayingPollInterval(interval time.Duration) Option {
	return func(srv *Server) { srv.nowPlayingPollInterval = interval }
}

// WithGitCommitsPollInterval sets the interval at which the server polls the
// GitCommitsService for updates.
func WithGitCommitsPollInterval(interval time.Duration) Option {
	return func(srv *Server) { srv.commitsPollInterval = interval }
}

// WithGitCommitsLimit configures the maximum number of Git commits to preload.
func WithGitCommitsLimit(limit int) Option {
	return func(srv *Server) { srv.commitsLimit = &limit }
}

// WithLogger cofnigures the Server to write logs using log.
func WithLogger(log *logrus.Logger) Option {
	return func(srv *Server) { srv.log = log }
}

// WithRaven configures the Server to capture panic events with the provided
// Raven client.
func WithRaven(rc *raven.Client) Option {
	return func(srv *Server) { srv.echo.Use(SentryRecoverMiddleware(rc)) }
}
