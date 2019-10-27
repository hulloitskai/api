package auth // import "go.stevenxie.me/api/auth"

import (
	"context"
	stderrs "errors"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/exthttp"
)

// Service is responsible for checking user permissions.
type Service interface {
	// GetPermissions gets the permissions granted to code.
	GetPermissions(ctx context.Context, code string) ([]Permission, error)

	// HasPermission returns true if code has the specified permission.
	HasPermission(
		ctx context.Context,
		code string, p Permission,
	) (ok bool, err error)
}

// ErrInvalidCode is returned by a Service when a provided code is invalid.
var ErrInvalidCode = exthttp.WrapWithHTTPCode(
	errors.WithDetailf(
		stderrs.New("auth: invalid code"),
		"Code is invalid or expired.",
	),
	http.StatusUnauthorized,
)
