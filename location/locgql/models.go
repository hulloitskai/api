package locgql

import (
	"go.stevenxie.me/api/location"
	"go.stevenxie.me/api/pkg/timeutil"
)

// A HistorySegment represents a segment of location history.
type HistorySegment struct {
	*location.HistorySegment
	Address  *string `json:"address,omitempty"`
	Distance *int    `json:"distance,omitempty"`
}

// An Address describes the position of a Place.
type Address struct {
	*location.Address
	District *string `json:"district,omitempty"`
	Street   *string `json:"street,omitempty"`
	Number   *string `json:"number,omitempty"`
}

// A Place is a location.Place that can represents its time zone as a string.
type Place struct {
	*location.Place
	Address Address
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
