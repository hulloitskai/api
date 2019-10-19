package scheduling // import "go.stevenxie.me/api/scheduling"

import (
	"context"
	"time"
)

// A Service provides scheduling information, derived from my calendar events
// and current location.
type Service interface {
	Service()
	BusyPeriods(ctx context.Context, date time.Time) ([]TimePeriod, error)
}
