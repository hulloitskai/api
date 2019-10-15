package authutil

import (
	stderrs "errors"
)

// ErrAccessDenied is a utility error that signifies that one does not have
// permission to use a resources.
var ErrAccessDenied = stderrs.New("authutil: access denied")
