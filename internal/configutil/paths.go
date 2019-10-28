package configutil

import (
	"fmt"

	"go.stevenxie.me/api/v2/internal"
)

// ConfigDir is the configuration directory corresponding to this module.
const ConfigDir = "/etc/" + internal.Namespace

// ConfigPaths returns a slice of config file locations for a given component
// name.
func ConfigPaths(component string) []string {
	return []string{
		component + ".yaml",
		fmt.Sprintf("%s/%s.yaml", ConfigDir, component),
	}
}
