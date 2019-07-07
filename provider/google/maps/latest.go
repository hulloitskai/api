package maps

import (
	"time"

	"github.com/stevenxie/api/pkg/geo"
)

// LastSegment returns the authenticated user's latest location history segment.
func (h *Historian) LastSegment() (*geo.Segment, error) {
	segments, err := h.LocationHistory(time.Now())
	if err != nil {
		return nil, err
	}
	if len(segments) == 0 {
		return nil, nil
	}
	return segments[len(segments)-1], nil
}
