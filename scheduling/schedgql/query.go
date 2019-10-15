package schedgql

import (
	"context"
	"time"

	"go.stevenxie.me/api/scheduling"
)

// NewQuery creates a new Query.
func NewQuery(svc scheduling.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for my scheduling-related data.
type Query struct {
	svc scheduling.Service
}

// BusyPeriods looks up the time periods when I'm busy.
func (q Query) BusyPeriods(
	ctx context.Context,
	date *time.Time,
) ([]scheduling.TimePeriod, error) {
	d := time.Now()
	if date != nil {
		d = (*date)
	}
	return q.svc.BusyPeriods(ctx, d)
}
