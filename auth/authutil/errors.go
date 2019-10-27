package authutil

import (
	stderrs "errors"
	"net/http"

	"github.com/cockroachdb/errors/exthttp"
)

// ErrAccessDenied is a utility error that signifies that one does not have
// permission to use a resources.
var ErrAccessDenied = exthttp.WrapWithHTTPCode(
	stderrs.New("authutil: access denied"),
	http.StatusUnauthorized,
)
