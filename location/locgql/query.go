package locgql

import (
	"context"
	"strings"
	"time"

	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/auth/authutil"

	"github.com/cockroachdb/errors"

	"github.com/99designs/gqlgen/graphql"
	funk "github.com/thoas/go-funk"
	"go.stevenxie.me/api/location"
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
func (q Query) Region(ctx context.Context) (*Place, error) {
	p, err := q.svc.CurrentRegion(
		ctx,
		func(cfg *location.CurrentRegionConfig) {
			fields := graphql.CollectAllFields(ctx)
			if funk.ContainsString(fields, "timeZone") {
				cfg.IncludeTimeZone = true
			}
		},
	)
	if err != nil {
		return nil, err
	}
	return &Place{p}, nil
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
