package config

import (
	"go.stevenxie.me/api/internal/info"
	"go.stevenxie.me/api/pkg/configutil"
)

// DefaultFilepaths are paths to look for config files.
var DefaultFilepaths = []string{
	info.Namespace + ".yaml",
	"/etc/" + info.Namespace + "/config.yaml",
}

// Load attempts to load a Config from its default filepaths.
func Load() (*Config, error) { return LoadFrom(DefaultFilepaths...) }

// LoadFrom attempts to the load a Config by checking for files located at
// filepaths.
func LoadFrom(filepaths ...string) (*Config, error) {
	cfg := defaultConfig()
	if err := configutil.TryLoadFiles(cfg, filepaths...); err != nil {
		return nil, err
	}
	return cfg, nil
}
