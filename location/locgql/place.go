package locgql

import (
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/timeutil"
)

// A Place is a location.Place that can represents its time zone as a string.
type Place struct {
	*location.Place
}

// A TimeZone represents a time zone.
type TimeZone struct {
	Name string `json:"name"`

	// The offset in seconds east of UTC.
	Offset int `json:"offset"`
}

// TimeZone returns the name of the Place's time zone.
func (p Place) TimeZone() *TimeZone {
	if p.Place.TimeZone == nil {
		return nil
	}
	name, offset := timeutil.CurrentZone(p.Place.TimeZone)
	return &TimeZone{
		Name:   name,
		Offset: offset,
	}
}
