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
	pairs := make([]string, 0, len(numberWords)*4)
	for word, n := range numberWords {
		nstr := strconv.FormatInt(int64(n), 10)
		pairs = append(pairs, word, nstr, strings.Title(word), nstr)
	}
	return strings.NewReplacer(pairs...).Replace(s)
}
