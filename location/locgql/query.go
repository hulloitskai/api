package locgql

import (
	"context"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cockroachdb/errors"
	funk "github.com/thoas/go-funk"

	"go.stevenxie.me/api/v2/auth"
	"go.stevenxie.me/api/v2/auth/authutil"
	"go.stevenxie.me/api/v2/location"
)

// NewQuery creates a new Query.
func NewQuery(svc location.Service, auth auth.Service) Query {
	return Query{
		svc:  svc,
		auth: auth,
	}
}

// A Query resolves queries for my music-related data.
type Query struct {
	svc  location.Service
	auth auth.Service
}

// Region resolves queries my current region.
func (q Query) Region(ctx context.Context) (*location.Place, error) {
	return q.svc.CurrentRegion(
		ctx,
		func(opt *location.CurrentRegionOptions) {
			fields := graphql.CollectAllFields(ctx)
			if funk.ContainsString(fields, "timeZone") {
				opt.IncludeTimeZone = true
			}
		},
	)
}

// History resolves queries for my location history.
func (q Query) History(
	ctx context.Context,
	code string,
	date *time.Time,
) ([]location.HistorySegment, error) {
	ok, err := q.auth.HasPermission(
		ctx,
		strings.TrimSpace(code), location.PermHistory,
	)
	if err != nil {
		return nil, errors.Wrap(err, "locgql: checking permissions")
	}
	if !ok {
		return nil, authutil.ErrAccessDenied
	}

	if date != nil {
		return q.svc.GetHistory(ctx, *date)
	}
	return q.svc.RecentHistory(ctx)
}
