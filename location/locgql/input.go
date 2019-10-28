package locgql

import "go.stevenxie.me/api/v2/location"

// A CoordinatesInput is a location.Coordinates with an optional Y component.
type CoordinatesInput struct {
	X float64  `json:"x"`
	Y float64  `jsom:"y"`
	Z *float64 `json:"z"`
}

// CoordinatesFromInput convertes a CoordinatesInput into a
// location.Coordinates.
func CoordinatesFromInput(ci CoordinatesInput) location.Coordinates {
	coords := location.Coordinates{
		X: ci.X,
		Y: ci.Y,
	}
	if ci.Z != nil {
		coords.Z = *ci.Z
	}
	return coords
}
