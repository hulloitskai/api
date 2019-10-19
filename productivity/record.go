package productivity

import (
	"context"
	"time"
)

// Productivity is a measure of productivity for a given day.
type Productivity struct {
	Records []Record `json:"records"`

	// Score is a number between 0 and 100, computed as follows:
	// https://help.rescuetime.com/article/73-how-is-my-productivity-pulse-calculated
	Score *uint `json:"score,omitempty"`
}

// Record is a record of the time spent doing activities of a particular
// Category.
type Record struct {
	Category Category      `json:"category"`
	Duration time.Duration `json:"duration"`
}

// A RecordSource can load CategoryRecords for a given date.
type RecordSource interface {
	GetRecords(ctx context.Context, date time.Time) ([]Record, error)
}
