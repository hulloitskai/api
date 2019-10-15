package locsvc

import "go.stevenxie.me/api/location"

// Get the latest coordinates in a HistorySegment, or nil if the HistorySegment
// contains no coordinates.
func latestCoordinates(seg *location.HistorySegment) *location.Coordinates {
	if n := len(seg.Coordinates); n > 0 {
		return &seg.Coordinates[n-1]
	}
	return nil
}
