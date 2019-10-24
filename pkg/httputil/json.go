package httputil

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.stevenxie.me/gopkg/zero"
)

const _jsonIndent = "  "

// JSONPretty sends v as JSON, but with pretty spacing.
func JSONPretty(c echo.Context, v zero.Interface) error {
	return c.JSONPretty(http.StatusOK, v, _jsonIndent)
}
