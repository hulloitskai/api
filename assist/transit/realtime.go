package transit

import (
	"context"
	stderrs "errors"
	"time"
)

// A RealtimeSource can get realtime transit departure information.
type RealtimeSource interface {
	GetDepartureTimes(context.Context, Transport, Station) ([]time.Time, error)
}

// ErrOperatorNotSupported reports that a particular operator is not supported
// for some operation.
var ErrOperatorNotSupported = stderrs.New("transit: operator not supported")
