package handler

import (
	echo "github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/api"
	"github.com/stevenxie/api/pkg/geo"
)

// LocationHandler handles requests for my whereabouts.
func LocationHandler(
	svc api.LocationService,
	log *logrus.Logger,
) echo.HandlerFunc {
	return func(c echo.Context) error {
		location, err := svc.CurrentRegion()
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
