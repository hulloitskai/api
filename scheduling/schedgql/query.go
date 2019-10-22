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

// BusyTimes looks up the times when I'm busy.
func (q Query) BusyTimes(
	ctx context.Context,
	date *time.Time,
) ([]scheduling.TimeSpan, error) {
	if date != nil {
		return q.svc.BusyTimes(ctx, *date)
	}
	return q.svc.BusyTimesToday(ctx)
}
