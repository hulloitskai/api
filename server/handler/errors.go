package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/stevenxie/api/pkg/errors"
)

type contextKey string

// StatusCodeContextKey is the context key that ErrorHandler will check to
// override the default error status code (http.StatusInternalServerError).
const StatusCodeContextKey contextKey = "StatusCode"

// ErrorHandler handles errors by writing them to c.Response() as JSON.
//
// It will attempt to extract the status code c.
func ErrorHandler(l zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var (
			data struct {
				Error string `json:"error"`
				Cause string `json:"cause,omitempty"`
				Code  *int   `json:"code,omitempty"`
			}
			statusCode = http.StatusInternalServerError
		)

		// Check if error comes from Echo.
		if herr, ok := err.(*echo.HTTPError); ok {
			statusCode = herr.Code
			data.Error = fmt.Sprint(herr.Message)
			goto send
		}

		// Retrieve status code from request context.
		if val := c.Request().Context().Value(StatusCodeContextKey); val != nil {
			code, ok := val.(int)
			if !ok {
				l.Error().Interface("contextVal", val).
					Msg("Unrecognized status code context value.")
				return // break early
			}
			statusCode = code
		}

		// Build error response.
		data.Error = err.Error()
		if wcode, ok := err.(errors.WithCode); ok { // check error code
			code := wcode.Code()
			data.Code = &code
		}
		if ecause := xerrors.Unwrap(err); ecause != nil { // check underlying cause
			data.Cause = ecause.Error()
		}

	send:
		// Send error, handle JSON marshalling failures.
		if err = c.JSONPretty(statusCode, &data, jsonPrettyIndent); err != nil {
			const msg = "Failed to write JSON error."
			c.Response().WriteHeader(http.StatusInternalServerError)
			io.WriteString(c.Response(), msg)
			l.Err(err).Msg(msg)
		}
	}
}

// setRequestStatusCode inserts the status code 'code' into the request context,
// tor use with ErrorHandler.
func setRequestStatusCode(c echo.Context, code int) {
	var (
		req  = c.Request()
		ctx  = context.WithValue(req.Context(), StatusCodeContextKey, code)
		nreq = req.WithContext(ctx)
	)
	c.SetRequest(nreq)
}
