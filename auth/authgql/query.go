package authgql

import (
	"context"
	"strings"

	"go.stevenxie.me/api/v2/auth"
)

// NewQuery creates a new Query.
func NewQuery(svc auth.Service) Query {
	return Query{svc: svc}
}

// A Query resolves queries for my auth-related data.
type Query struct {
	svc auth.Service
}

// Permissions gets the permissions granted to code.
func (q Query) Permissions(ctx context.Context, code string) (
	perms []string, err error) {
	ps, err := q.svc.GetPermissions(ctx, strings.TrimSpace(code))
	if err != nil {
		return nil, err
	}
	perms = make([]string, len(ps))
	for i, p := range ps {
		perms[i] = string(p)
	}
	return perms, nil
}
