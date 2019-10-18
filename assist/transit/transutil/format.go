package transutil

import (
	"regexp"
	"strings"
)

var matchAmpWithoutSpace = regexp.MustCompile(`([^ ])&([^ ])`)

func fixAmpSpacing(s string) string {
	return matchAmpWithoutSpace.ReplaceAllString(s, "$1 & $2")
}

// NormalizeStationName normalizes a station name.
func NormalizeStationName(s string) string {
	s = strings.Trim(s, ` /()`)
	s = strings.ReplaceAll(s, "/", "&")
	s = fixAmpSpacing(s)
	s = strings.ReplaceAll(s, "Of", "of")
	return s
}
