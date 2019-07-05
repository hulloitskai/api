package server

import (
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
)

// WithLogger cofnigures the Server to write logs using log.
func WithLogger(log *logrus.Logger) Option {
	return func(srv *Server) { srv.log = log }
}

// WithRaven configures the Server to capture panic events with the provided
// Raven client.
func WithRaven(rc *raven.Client) Option {
	return func(srv *Server) { srv.echo.Use(SentryRecoverMiddleware(rc)) }
}
