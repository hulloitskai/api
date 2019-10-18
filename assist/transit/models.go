package transit

import (
	"time"

	"go.stevenxie.me/api/location"
)

// A Station is a place where one can board a transit vehicle.
type Station struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Coordinates location.Coordinates `json:"coordinates"`
}

// A Transport is a vehicle travelling on a transit route.
type Transport struct {
	Route     string    `json:"route"`
	Direction string    `json:"direction"`
	Category  string    `json:"category"`
	Operator  *Operator `json:"operator"`
}

// An Operator represents a transit system operator.
type Operator struct {
	Code string `json:"string"`
	Name string `json:"name"`
}

// A Departure represents the departure of a Transport from a Station.
type Departure struct {
	Times     []time.Time `json:"times"`
	Transport *Transport  `json:"transport"`
	Station   *Station    `json:"station"`
	Realtime  bool        `json:"realtime"`
}

// RelativeTimes returns the departure times as a time.Duration relative to
// the current time.
func (d *Departure) RelativeTimes() []time.Duration {
	var (
		rt  = make([]time.Duration, len(d.Times))
		now = time.Now()
	)
	for i, t := range d.Times {
		rt[i] = t.Sub(now)
	}
	return rt
}
