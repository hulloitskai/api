package location

import (
	"context"
	"fmt"
	"time"

	"go.stevenxie.me/api/scheduling"
)

// A HistorySegment represents a segment of location history.
type HistorySegment struct {
	Place       string              `json:"place"`
	Address     string              `json:"address,omitempty"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Distance    int                 `json:"distance,omitempty"`
	TimeSpan    scheduling.TimeSpan `json:"timeSpan"`
	Coordinates []Coordinates       `json:"coordinates"`
}

func (seg *HistorySegment) String() string {
	return fmt.Sprintf("%+v", *seg)
}

// A Historian can get my location history segments for a particular
// date.
type Historian interface {
	GetHistory(ctx context.Context, date time.Time) ([]HistorySegment, error)
}
