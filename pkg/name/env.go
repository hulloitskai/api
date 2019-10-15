package name

import "strings"

// EnvKey derives the name of the environment key for a set of strings (which
// will be joined by underscores).
func EnvKey(parts ...string) string {
	for i, part := range parts {
		parts[i] = strings.ToUpper(part)
	}
	return strings.Join(parts, "_")
}
