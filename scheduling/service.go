package scheduling // import "go.stevenxie.me/api/scheduling"

import (
	"context"
	"time"
)

// A Service provides scheduling information, derived from my calendar events
// and current location.
type Service interface {
	// BusyTimes gets my busy times for the given date, sorted in ascending order
	// by start time.
	BusyTimes(ctx context.Context, date time.Time) ([]TimeSpan, error)

	BusyTimesToday(ctx context.Context) ([]TimeSpan, error)
}
