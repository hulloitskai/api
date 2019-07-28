package httputil

import (
	"net/http"
	"os"

	echo "github.com/labstack/echo/v4"
)

// InfoHandler creates an echo.HandlerFunc that handles requests for server
// information.
func InfoHandler(name, version string) echo.HandlerFunc {
	return func(c echo.Context) error {
		data := struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Environment string `json:"environment,omitempty"`
		}{
			Name:        name,
			Version:     version,
			Environment: os.Getenv("GOENV"),
		}
		return c.JSONPretty(http.StatusOK, data, "  ")
	}
}
