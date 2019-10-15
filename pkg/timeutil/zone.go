package timeutil

import "time"

// CurrentZone computes the current effective time zone of l.
func CurrentZone(l *time.Location) (name string, offset int) {
	now := time.Now()
	return now.In(l).Zone()
}
