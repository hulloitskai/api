package productivity // import "go.stevenxie.me/api/v2/productivity"

import (
	"context"
	"time"
)

// A Service provides productivity metrics information.
type Service interface {
	GetProductivity(ctx context.Context, date time.Time) (*Productivity, error)
	CurrentProductivity(ctx context.Context) (*Productivity, error)
}
