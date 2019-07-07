package handler

import (
	"net/http"
	"strings"

	errors "golang.org/x/xerrors"

	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/geo"
	"github.com/stevenxie/api/pkg/httputil"
)

// A LocationProvider can create handlers that use location data.
type LocationProvider struct{ svc api.LocationService }

// NewLocationProvider creates a new LocationProvider.
func NewLocationProvider(svc api.LocationService) LocationProvider {
	return LocationProvider{svc}
}

// RegionHandler handles requests for my current geographical region.
func (p LocationProvider) RegionHandler(log *logrus.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		location, err := p.svc.CurrentRegion()
		if err != nil {
			log.WithError(err).Error("Failed to get current region.")
			return err
		}

		data := struct {
			*geo.Location
			Shape [][]float64 `json:"shape"`
		}{
			Location: location,
			Shape:    make([][]float64, len(location.Shape)),
		}
		for i, shape := range data.Location.Shape {
			data.Shape[i] = []float64{shape.X, shape.Y}
		}
		return jsonPretty(c, &data)
	}
}

const bearerTokenPrefix = "Bearer "

// RecentHistoryHandler handles requests for my recent location history.
func (p LocationProvider) RecentHistoryHandler(
	access api.LocationAccessService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse bearer token as access code..
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			httputil.SetEchoStatusCode(c, http.StatusUnauthorized)
			return errors.New("no authorization header")
		}
		if !strings.HasPrefix(authHeader, bearerTokenPrefix) {
			httputil.SetEchoStatusCode(c, http.StatusBadRequest)
			return errors.New("invalid authorization header: invalid bearer token " +
				"format")
		}
		token := strings.TrimPrefix(authHeader, bearerTokenPrefix)

		// Validate access code.
		valid, err := access.IsValidCode(token)
		if err != nil {
			log.WithError(err).Error("Failed to validate access token.")
			return errors.Errorf("failed to validate access code: %w", err)
		}
		if !valid {
			httputil.SetEchoStatusCode(c, http.StatusUnauthorized)
			return errors.New("access code is invalid or expired")
		}

		// Retrieve last location history segment.
		segment, err := p.svc.LastSegment()
		if err != nil {
			log.WithError(err).Error("Failed to get last position.")
			return errors.Errorf("failed to get last position: %w", err)
		}
		if segment == nil {
			return errors.New("no location history segments found")
		}

		coordinates := make([][]float64, len(segment.Coordinates))
		for i, coord := range segment.Coordinates {
			coordinates[i] = []float64{coord.X, coord.Y}
		}

		data := struct {
			*geo.Segment
			Coordinates [][]float64 `json:"coordinates"`
		}{
			Segment:     segment,
			Coordinates: coordinates,
		}
		return jsonPretty(c, data)
	}
}
