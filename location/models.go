package location

import "time"

type (
	// A Place is a geographical location.
	Place struct {
		ID       string         `json:"id"`
		Level    string         `json:"level"`
		Type     string         `json:"type"`
		Position Coordinates    `json:"position"`
		TimeZone *time.Location `json:"timeZone,omitempty"`
		Address  Address        `json:"address"`
		Shape    []Coordinates  `json:"shape,omitempty"`
	}

	// Coordinates represents a point in 3D space.
	Coordinates struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	}

	// An Address describes the position of a Place.
	Address struct {
		Label    string `json:"label"`
		Country  string `json:"country"`
		State    string `json:"state"`
		County   string `json:"county"`
		City     string `json:"city"`
		District string `json:"district,omitempty"`
		Postcode string `json:"postcode"`
		Street   string `json:"street,omitempty"`
		Number   string `json:"number,omitempty"`
	}
)
