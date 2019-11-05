package httputil

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/exthttp"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// ErrorHandler creates an echo.HTTPErrorHandler that handles errors by
// sending JSON responses containing error details.
//
// It will attempt to extract additional error context using cockroachdb/errors.
func ErrorHandler(log *logrus.Entry) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		var data struct {
			Error   string   `json:"error"`
			Cause   string   `json:"cause,omitempty"`
			Details []string `json:"details,omitempty"`
			Hints   []string `json:"hints,omitempty"`
		}

		// Check if error contains a status code.
		status := exthttp.GetHTTPCode(err, http.StatusInternalServerError)

		// Check if error comes from Echo.
		{
			var httpError *echo.HTTPError
			if errors.As(err, &httpError) {
				status = httpError.Code
				data.Error = strings.ToLower(fmt.Sprint(httpError.Message))
				goto Send
			}
		}

		// Build error response.
		data.Error = err.Error()
		if cause := errors.UnwrapAll(err); !errors.Is(cause, err) {
			if cause.Error() != err.Error() {
				data.Cause = cause.Error()
			}
		}
		if details := errors.GetAllDetails(err); len(details) > 0 {
			data.Details = details
		}
		if hints := errors.GetAllHints(err); len(hints) > 0 {
			data.Hints = hints
		}

	Send:
		// Send error, handle JSON marshaling failures.
		if err = c.JSON(status, &data); err != nil {
			const msg = "Failed to write JSON error."
			c.String(http.StatusInternalServerError, msg)
			log.WithError(err).Error(msg)
		}
	}
}
