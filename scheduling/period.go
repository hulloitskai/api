package scheduling

import (
	"encoding/json"
	"time"
)

// TimeLayout is the layout of time used to encode and decode the time.Times in
// a TimePeriod.
const TimeLayout = time.RFC3339

// TimePeriod represents a period of time.
type TimePeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Before reports whether the time period tp is before up.
func (tp *TimePeriod) Before(up *TimePeriod) bool {
	if tp.Start.Before(up.Start) {
		return true
	}
	if tp.Start.Equal(up.Start) {
		return tp.End.Before(up.End)
	}
	return false
}

// Equal reports whether the time period tp is equal to up.
func (tp *TimePeriod) Equal(up *TimePeriod) bool {
	return tp.Start.Equal(up.Start) && tp.End.Equal(up.End)
}

// After reports whether the time period tp is after up.
func (tp *TimePeriod) After(up *TimePeriod) bool {
	if tp.Start.After(up.End) {
		return true
	}
	if tp.Start.Equal(up.Start) {
		return tp.End.After(up.End)
	}
	return false
}

// MarshalJSON implements a json.Marshaller for a Period.
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
