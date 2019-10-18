package transit

import (
	"context"
	stderrs "errors"
	"time"
)

// A RealTimeService can get realtime transit departure information.
type RealTimeService interface {
	GetDepartureTimes(context.Context, Transport, Station) ([]time.Time, error)
}

// ErrOperatorNotSupported reports that a particular operator is not supported
// for some operation.
var ErrOperatorNotSupported = stderrs.New("transit: operator not supported")
