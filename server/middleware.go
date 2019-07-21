package server

import (
	"github.com/getsentry/raven-go"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// SentryRecoverMiddleware is an echo.MiddlewareFunc that captures panics and
// reports them to sentry.
func SentryRecoverMiddleware(
	rc *raven.Client,
	log *logrus.Logger,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			if v, id := rc.CapturePanic(func() {
				err = next(c)
			}, nil); v != nil {
				log.
					WithField("id", id).
					Warn("A handler panic was captured by Sentry.")
				log.Panic(v)
			}
			return err
		}
	}
}
