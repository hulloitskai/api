package scheduling

import (
	"context"
	"time"
)

type (
	// A Service provides scheduling information, derived from my calendar events
	// and current location.
	Service interface {
		BusyPeriods(ctx context.Context, date time.Time) ([]TimePeriod, error)
	}

	// A BusySource can determine my busy time periods.
	BusySource interface {
		BusyPeriods(ctx context.Context, date time.Time) ([]TimePeriod, error)
	}
)
