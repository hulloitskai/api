package schedgql

import (
	"context"
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/auth/authutil"
	"go.stevenxie.me/api/pkg/timeutil"

	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/scheduling"
)

// NewQuery creates a new Query.
func NewQuery(svc scheduling.Service, auth auth.Service) Query {
	return Query{
		svc:  svc,
		auth: auth,
	}
}

// A Query resolves queries for my scheduling-related data.
type Query struct {
	svc  scheduling.Service
	auth auth.Service
}

// BusyTimes looks up the times when I'm busy.
func (q Query) BusyTimes(
	ctx context.Context,
	code *string,
	date *time.Time,
) ([]scheduling.TimeSpan, error) {
	if date != nil {
		// If date is before or after today, only allow access to users with
		// scheduling.PermBusyAll.
		var (
			start = timeutil.DayStart(time.Now())
			end   = start.AddDate(0, 0, 1)
		)
		if date.Before(start) || date.After(end) {
			if code == nil {
				return nil, errors.WithDetail(
					authutil.ErrAccessDenied,
					"No code was provided.",
				)
			}
			ok, err := q.auth.HasPermission(ctx, *code, scheduling.PermBusyAll)
			if err != nil {
				return nil, errors.Wrap(err, "checking permissions")
			}
			if !ok {
				return nil, authutil.ErrAccessDenied
			}
		}
		return q.svc.BusyTimes(ctx, *date)
	}
	return q.svc.BusyTimesToday(ctx)
}
