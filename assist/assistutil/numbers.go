package assistutil

import (
	"strconv"
	"strings"
)

var numberWords = map[string]int{
	"zero":  0,
	"one":   1,
	"two":   2,
	"three": 3,
	"four":  4,
	"five":  5,
	"six":   6,
	"seven": 7,
	"eight": 8,
	"nine":  9,
}

// ReplaceNumberWords replaces the number words in s with their actual integer
// values; especially useful for parsing Siri dictation.
//
// Only works with number words between 0 and 9.
func ReplaceNumberWords(s string) string {
	// Try replacing single-word s.
	sTrimmed := strings.TrimSpace(s)
	for k, v := range numberWords {
		for _, variant := range []string{k, strings.Title(k)} {
			if variant == sTrimmed {
				return strings.ReplaceAll(s, variant, strconv.FormatInt(int64(v), 10))
			}
		}
	}

	// Try replacing in multi-word s.
	pairs := make([]string, 0, len(numberWords)*6)
	for word, n := range numberWords {
		nstr := strconv.FormatInt(int64(n), 10)
		pairs = append(
			pairs,
			word+" ", nstr+" ",
			" "+word, " "+nstr,
			strings.Title(word), nstr+" ",
		)
	}
	return strings.NewReplacer(pairs...).Replace(s)
}
