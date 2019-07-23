package handler

import (
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/pkg/httputil"
	"go.stevenxie.me/api/service/location"
)

// A LocationProvider can create handlers that use location data.
type LocationProvider struct {
	svc location.Service
}

// NewLocationProvider creates a new LocationProvider.
func NewLocationProvider(svc location.Service) LocationProvider {
	return LocationProvider{svc}
}

// CurrentRegionHandler handles requests for my current geographical region.
func (p LocationProvider) CurrentRegionHandler(
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		place, err := p.svc.CurrentRegion()
		if err != nil {
			log.WithError(err).Error("Failed to get current region.")
			return err
		}

		var (
			pos  = &place.Position
			data = struct {
				*location.Place
				Position []float64   `json:"position"`
				Shape    [][]float64 `json:"shape"`
			}{
				Place:    place,
				Position: []float64{pos.X, pos.Y},
				Shape:    make([][]float64, len(place.Shape)),
			}
		)
		for i, coords := range data.Place.Shape {
			data.Shape[i] = []float64{coords.X, coords.Y}
		}
		return jsonPretty(c, &data)
	}
}

const bearerTokenPrefix = "Bearer "

// HistoryHandler handles requests for my recent location history.
func (p LocationProvider) HistoryHandler(
	access location.AccessService,
	log *logrus.Logger,
) echo.HandlerFunc {
	handler := func(c echo.Context) error {
		// Retrieve recent location history segments.
		segments, err := p.svc.RecentHistory()
		if err != nil {
			log.WithError(err).Error("Failed to get recent location history.")
			return errors.Wrap(err, "failed to get recent location history")
		}

		type segment struct {
			*location.HistorySegment
			Coordinates [][]float64 `json:"coordinates"`
		}
		results := make([]segment, len(segments))
		for i, seg := range segments {
			coordinates := make([][]float64, len(seg.Coordinates))
			for j, coord := range seg.Coordinates {
				coordinates[j] = []float64{coord.X, coord.Y}
			}
			results[i] = segment{
				HistorySegment: seg,
				Coordinates:    coordinates,
			}
		}

		return jsonPretty(c, results)
	}

	return locationAccessValidationMiddlware(access, log)(handler)
}

func locationAccessValidationMiddlware(
	svc location.AccessService,
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
