package timeutil

import "time"

// DayStart computes a time.Time corresponding to the beginning of the day
// in which t takes place.
func DayStart(t time.Time) time.Time {
	return time.Date(
		t.Year(), t.Month(), t.Day(),
		0, 0, 0, 0,
		t.Location(),
	)
}
