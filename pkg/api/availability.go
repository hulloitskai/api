package api

import (
	"encoding/json"
	"time"
)

// TimeLayout is the layout of time used to encode and decode the time.Times in
// a TimePeriod.
const TimeLayout = time.RFC3339

// TimePeriod represents a block of time.
type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

//revive:disable
func (tp *TimePeriod) MarshalJSON() ([]byte, error) {
	var formatted = struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}{
		Start: tp.Start.Format(TimeLayout),
		End:   tp.End.Format(TimeLayout),
	}
	return json.Marshal(&formatted)
}

// An AvailabilityService is able to fetch periods of availability from a
// calendar.
type AvailabilityService interface {
	// BusyPeriods returns the busy periods for a given date.
	BusyPeriods(date time.Time) ([]*TimePeriod, error)
	// Timezone returns my timezone.
	Timezone() (*time.Location, error)
}
