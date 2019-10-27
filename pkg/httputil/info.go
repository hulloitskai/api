package httputil

import (
	"encoding/json"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.stevenxie.me/gopkg/configutil"
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
			Environment: configutil.GetGoEnv(),
		}
		return c.JSONPretty(http.StatusOK, data, "  ")
	}
}

// InfoHTTPHandler creates an http.HandlerFunc that handles requests for server
// information.
func InfoHTTPHandler(name, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			Environment string `json:"environment,omitempty"`
		}{
			Name:        name,
			Version:     version,
			Environment: configutil.GetGoEnv(),
		}
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if err := enc.Encode(&data); err != nil {
			panic(err)
		}
	}
}
