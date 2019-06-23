package handler

import (
	"os"

	echo "github.com/labstack/echo/v4"
	"github.com/stevenxie/api/internal/info"
	srvinfo "github.com/stevenxie/api/server/internal/info"
)

// InfoHandler handles requests for information about the API server.
func InfoHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Environment string `json:"environment,omitempty"`
		}{
			Name:        srvinfo.Name,
			Version:     info.Version,
			Environment: os.Getenv("GOENV"),
		}
		return jsonPretty(c, &data)
	}
}
