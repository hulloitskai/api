package handler

import (
	"fmt"
	"io"
	"net/http"

	"golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/internal/httputil"
	"github.com/stevenxie/api/pkg/errors"
)

// ErrorHandler handles errors by writing them to c.Response() as JSON.
//
// It will attempt to extract the status code c.
func ErrorHandler(log *logrus.Logger) echo.HTTPErrorHandler {
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
		if code, ok := httputil.GetEchoStatusCode(c); ok {
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
			log.WithError(err).Error(msg)
		}
	}
}
