package httputil

import (
	"context"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

type contextKey string

// StatusCodeContextKey is the context key that describes the status code
// of a request response.
const StatusCodeContextKey contextKey = "StatusCode"

// SetStatusCode inserts the status code 'code' into the request.
func SetStatusCode(req *http.Request, code int) *http.Request {
	ctx := context.WithValue(req.Context(), StatusCodeContextKey, code)
	return req.WithContext(ctx)
}

// SetEchoStatusCode sets the status code of an echo.Context.
func SetEchoStatusCode(c echo.Context, code int) {
	c.SetRequest(SetStatusCode(c.Request(), code))
}

// GetStatusCode returns the status code embedded in the request, if any.
func GetStatusCode(req *http.Request) (code int, ok bool) {
	val := req.Context().Value(StatusCodeContextKey)
	if val == nil {
		return 0, false
	}
	return val.(int), true
}

// GetEchoStatusCode gets the status code of an echo.Context.
func GetEchoStatusCode(c echo.Context) (code int, ok bool) {
	return GetStatusCode(c.Request())
}
