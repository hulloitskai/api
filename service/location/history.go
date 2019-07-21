package location

import "time"

// A HistoryService can fetch location history segments.
type HistoryService interface {
	RecentHistory() (HistorySegments, error)
}

type (
	// A HistorySegment represents a location history segment.
	HistorySegment struct {
		Place       string        `json:"place"`
		Address     string        `json:"address,omitempty"`
		Description string        `json:"description"`
		Category    string        `json:"category"`
		Distance    int           `json:"distance,omitempty"`
		TimeSpan    TimeSpan      `json:"timeSpan"`
		Coordinates []Coordinates `json:"coordinates"`
	}

	// HistorySegments are a set of HistorySegments.
	HistorySegments []*HistorySegment

	// TimeSpan is a span of time.
	TimeSpan struct {
		Begin time.Time `json:"begin"`
		End   time.Time `json:"end"`
	}
)
