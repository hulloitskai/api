package util

import (
	"go.uber.org/zap"
)

// Noop does nothing ¯\_(ツ)_/¯.
func Noop() {}

// Empty contains no information.
type Empty struct{}

// NoopLogger is a no-op logger that does not write any logs.
var NoopLogger = zap.NewNop().Sugar()
