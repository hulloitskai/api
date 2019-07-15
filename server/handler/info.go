package handler

import (
	"os"

	echo "github.com/labstack/echo/v4"
	"github.com/stevenxie/api/internal/info"
	serverinfo "github.com/stevenxie/api/server/internal/info"
)

// InfoHandler handles requests for information about the API server.
func InfoHandler() echo.HandlerFunc {
	return func(c echo.Context) error {
		data := struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Environment string `json:"environment,omitempty"`
		}{
			Name:        serverinfo.Name,
			Version:     info.Version,
			Environment: os.Getenv("GOENV"),
		}
		return jsonPretty(c, &data)
	}
}
