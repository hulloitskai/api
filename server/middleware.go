package server

import (
	"github.com/getsentry/raven-go"
	echo "github.com/labstack/echo/v4"
)

// SentryRecoverMiddleware is an echo.MiddlewareFunc that captures panics and
// reports them to sentry.
func SentryRecoverMiddleware(rc *raven.Client) echo.MiddlewareFunc {
	return func(handler echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var err error
			rc.CapturePanic(func() { err = handler(c) }, nil)
			return err
		}
	}
}
