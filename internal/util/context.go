package util

import (
	"context"
	"time"
)

// ContextWithTimeout produces a context with a timeout if timeout > 0, and
// a background context otherwise (with a noop context.CancelFunc).
func ContextWithTimeout(timeout time.Duration) (context.Context,
	context.CancelFunc) {
	bg := context.Background()
	if timeout > 0 {
		return context.WithTimeout(bg, timeout)
	}
	return bg, Noop
}
