package prodgql

import (
	"context"

	"go.stevenxie.me/api/v2/productivity"
)

// NewQuery creates a new Query.
func NewQuery(svc productivity.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for my personal information.
type Query struct {
	svc productivity.Service
}

//revive:disable-line:exported
func (q Query) Productivity(ctx context.Context) (*productivity.Productivity, error) {
	return q.svc.CurrentProductivity(ctx)
}
