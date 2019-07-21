package availability

import (
	"encoding/json"
	"time"
)

// TimeLayout is the layout of time used to encode and decode the time.Times in
// a TimePeriod.
const TimeLayout = time.RFC3339

type (
	// Period represents a block of time.
	Period struct {
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// Periods are a set of Periods.
	Periods []*Period
)

//revive:disable
func (tp *Period) MarshalJSON() ([]byte, error) {
	var formatted = struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}{
		Start: tp.Start.Format(TimeLayout),
		End:   tp.End.Format(TimeLayout),
	}
	return json.Marshal(&formatted)
}

// A Service is able to fetch periods of availability, and determine my
// timezone.
type Service interface {
	// BusyPeriods returns the busy periods for a given date.
	BusyPeriods(date time.Time) (Periods, error)

	// Timezone returns my timezone.
	Timezone() (*time.Location, error)
}
