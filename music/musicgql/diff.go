package musicgql

import "go.stevenxie.me/api/v2/music"

// IsEqualsCurrentlyPlaying returns true if x is equal to y.
func IsEqualsCurrentlyPlaying(x, y *music.CurrentlyPlaying) bool {
	if (x == nil) || (y == nil) {
		return x == y
	}
	if x.Playing != y.Playing {
		return false
	}
	if x.Progress != y.Progress {
		return false
	}
	if x.Track.ID != y.Track.ID {
		return false
	}
	return true
}
