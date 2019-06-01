package api

import "time"

// TimePeriod represents a block of time.
type TimePeriod struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// An AvailabilityService is able to fetch periods of availability from a
// calendar.
type AvailabilityService interface {
	// BusyPeriods returns the busy periods for a given date.
	BusyPeriods(date time.Time) ([]*TimePeriod, error)
	// Timezone returns my timezone.
	Timezone() (*time.Location, error)
}
