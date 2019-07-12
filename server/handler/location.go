package handler

import (
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
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

// CurrentRegionHandler handles requests for my current geographical region.
func (p LocationProvider) CurrentRegionHandler(
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		location, err := p.svc.CurrentRegion()
		if err != nil {
			log.WithError(err).Error("Failed to get current region.")
			return err
		}

		var (
			pos  = &location.Position
			data = struct {
				*geo.Location
				Position []float64   `json:"position"`
				Shape    [][]float64 `json:"shape"`
			}{
				Location: location,
				Position: []float64{pos.X, pos.Y},
				Shape:    make([][]float64, len(location.Shape)),
			}
		)
		for i, shape := range data.Location.Shape {
			data.Shape[i] = []float64{shape.X, shape.Y}
		}
		return jsonPretty(c, &data)
	}
}

const bearerTokenPrefix = "Bearer "

// HistoryHandler handles requests for my recent location history.
func (p LocationProvider) HistoryHandler(
	access api.LocationAccessService,
	log *logrus.Logger,
) echo.HandlerFunc {
	handler := func(c echo.Context) error {
		// Retrieve recent location history segments.
		segments, err := p.svc.RecentSegments()
		if err != nil {
			log.WithError(err).Error("Failed to get recent location history.")
			return errors.Wrap(err, "failed to get recent location history")
		}

		type segment struct {
			*geo.Segment
			Coordinates [][]float64 `json:"coordinates"`
		}
		results := make([]segment, len(segments))
		for i, seg := range segments {
			coordinates := make([][]float64, len(seg.Coordinates))
			for j, coord := range seg.Coordinates {
				coordinates[j] = []float64{coord.X, coord.Y}
			}
			results[i] = segment{
				Segment:     seg,
				Coordinates: coordinates,
			}
		}
		return jsonPretty(c, results)
	}

	return locationAccessValidationMiddlware(access, log)(handler)
}

func locationAccessValidationMiddlware(
	svc api.LocationAccessService,
	log *logrus.Logger,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Parse bearer token as access code.
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				httputil.SetEchoStatusCode(c, http.StatusUnauthorized)
				return errors.New("no authorization header")
			}
			if !strings.HasPrefix(authHeader, bearerTokenPrefix) {
				httputil.SetEchoStatusCode(c, http.StatusBadRequest)
				return errors.New(
					"invalid authorization header: invalid bearer token format",
				)
			}
			token := strings.TrimPrefix(authHeader, bearerTokenPrefix)

			// Validate access code.
			valid, err := svc.IsValidCode(token)
			if err != nil {
				log.WithError(err).Error("Failed to validate access token.")
				return errors.Wrap(err, "failed to validate access code")
			}
			if !valid {
				httputil.SetEchoStatusCode(c, http.StatusUnauthorized)
				return errors.New("access code is invalid or expired")
			}

			return next(c)
		}
	}
}
