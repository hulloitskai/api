package httputil

import (
	"strings"

	"github.com/labstack/echo/v4"

	errors "golang.org/x/xerrors"
)

// WithContentType is middleware that executes the next handler if the request
// has the specified Content-Type header.
func WithContentType(contentType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctype := c.Request().Header.Get("Content-Type")
			if !strings.HasPrefix(ctype, contentType) {
				return errors.Errorf("unexpected Content-Type header '%s'", ctype)
			}

			// Call next handler.
			return next(c)
		}
	}
}
