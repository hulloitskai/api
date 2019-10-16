package location

import (
	"fmt"
	"time"
)

type (
	// A HistorySegment represents a segment of location history.
	HistorySegment struct {
		Place       string        `json:"place"`
		Address     string        `json:"address,omitempty"`
		Description string        `json:"description"`
		Category    string        `json:"category"`
		Distance    int           `json:"distance,omitempty"`
		TimeSpan    TimeSpan      `json:"timeSpan"`
		Coordinates []Coordinates `json:"coordinates"`
	}

	// TimeSpan is a period of time.
	TimeSpan struct {
		Begin time.Time `json:"begin"`
		End   time.Time `json:"end"`
	}
)

func (seg *HistorySegment) String() string {
	return fmt.Sprintf("%+v", *seg)
}