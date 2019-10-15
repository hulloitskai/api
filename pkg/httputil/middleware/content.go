package middleware

import (
	stderrs "errors"
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/exthttp"
	echo "github.com/labstack/echo/v4"
)

// AllowContentType is a middleware that only executes the next handler if the
// request has a matching Content-Type header.
func AllowContentType(types ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			ctype := c.Request().Header.Get("Content-Type")
			if ctype == "" {
				err = ErrMissingContentTypeHeader
				goto Fail
			}
			for _, allowed := range types {
				if strings.HasPrefix(ctype, allowed) {
					return next(c)
				}
			}

			// Return an error if ctype is not in the whitelist.
			err = errors.Newf(
				"middleware: unexpected Content-Type header '%s'",
				ctype,
			)
			err = errors.WithDetailf(
				err,
				"Unexpected Content-Type header '%s'", ctype,
			)

		Fail:
			return exthttp.WrapWithHTTPCode(err, http.StatusBadRequest)
		}
	}
}

// ErrMissingContentTypeHeader is returned by AllowContentType when an
// incoming request does not contain a Content-Type MIME header.
var ErrMissingContentTypeHeader = stderrs.New("middleware: missing " +
	"Content-Type header")
