package geo

import "time"

type (
	// A Placemark is a place that one has visited.
	Placemark struct {
		Name        string `xml:"name"`
		Address     string `xml:"address"`
		Description string `xml:"description"`
		Category    string
		Distance    int
		TimeSpan    TimeSpan
		Coordinates []Coordinate
	}

	// A TimeSpan is a span of time.
	TimeSpan struct {
		Begin time.Time `xml:"begin"`
		End   time.Time `xml:"end"`
	}

	// A Coordinate represents a point in 3D space.
	Coordinate struct {
		X float64
		Y float64
		Z float64
	}
)
