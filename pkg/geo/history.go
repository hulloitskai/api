package geo

import "time"

// A LocationHistoryService can fetch location history segments.
type LocationHistoryService interface {
	LocationHistory(date time.Time) ([]*Segment, error)
}

type (
	// A Segment represents a location history segment.
	Segment struct {
		Place       string       `json:"place"`
		Address     string       `json:"address,omitempty"`
		Description string       `json:"description"`
		Category    string       `json:"category"`
		Distance    int          `json:"distance,omitempty"`
		TimeSpan    TimeSpan     `json:"timeSpan"`
		Coordinates []Coordinate `json:"coordinates"`
	}

	// TimeSpan is a span of time.
	TimeSpan struct {
		Begin time.Time `json:"begin"`
		End   time.Time `json:"end"`
	}
)
