package aboutgql

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"

	"go.stevenxie.me/api/about"
	"go.stevenxie.me/api/auth"
	"go.stevenxie.me/api/auth/authutil"
)

// NewQuery creates a new Query.
func NewQuery(svc about.Service, auth auth.Service) Query {
	return Query{
		svc:  svc,
		auth: auth,
	}
}

// A Query resolves queries for my personal information.
type Query struct {
	svc  about.Service
	auth auth.Service
}

// About resolves requests for my personal information.
func (q Query) About(
	ctx context.Context,
	code *string,
) (about.ContactInfo, error) {
	if code != nil {
		ok, err := q.auth.HasPermission(
			ctx,
			strings.TrimSpace(*code), about.PermFull,
		)
		if err != nil {
			return nil, errors.Wrap(err, "svcgql: checking permissions")
		}
		if !ok {
			return nil, authutil.ErrAccessDenied
		}
		return q.svc.GetAbout(ctx)
	}
	return q.svc.GetMasked(ctx)
}
