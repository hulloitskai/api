package handler

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

const jsonPrettyIndent = "  "

func jsonPretty(c echo.Context, v interface{}) error {
	return c.JSONPretty(http.StatusOK, v, jsonPrettyIndent)
}
