package scheduling

import (
	"context"
	"encoding/json"
	"time"
)

// TimeLayout is the layout of time used to encode and decode the time.Times in
// a TimePeriod.
const TimeLayout = time.RFC3339

// TimeSpan represents a span of time.
type TimeSpan struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Before reports whether the time period tp is before up.
func (ts *TimeSpan) Before(up *TimeSpan) bool {
	if ts.Start.Before(up.Start) {
		return true
	}
	if ts.Start.Equal(up.Start) {
		return ts.End.Before(up.End)
	}
	return false
}

// Equal reports whether the time period tp is equal to up.
func (ts *TimeSpan) Equal(up *TimeSpan) bool {
	return ts.Start.Equal(up.Start) && ts.End.Equal(up.End)
}

// After reports whether the time period tp is after up.
func (ts *TimeSpan) After(up *TimeSpan) bool {
	if ts.Start.After(up.End) {
		return true
	}
	if ts.Start.Equal(up.Start) {
		return ts.End.After(up.End)
	}
	return false
}

// MarshalJSON implements a json.Marshaller for a Period.
func (ts *TimeSpan) MarshalJSON() ([]byte, error) {
	var formatted = struct {
		Start string `json:"start"`
		End   string `json:"end"`
	}{
		Start: ts.Start.Format(TimeLayout),
		End:   ts.End.Format(TimeLayout),
	}
	return json.Marshal(&formatted)
}

// A Calendar can get my busy time periods.
type Calendar interface {
	RawBusyTimes(ctx context.Context, date time.Time) ([]TimeSpan, error)
}
