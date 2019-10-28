package config

import (
	"go.stevenxie.me/api/v2/cmd/server/internal"
	"go.stevenxie.me/api/v2/internal/configutil"
)

// Load finds and loads a Config from a set of possible file locations.
func Load(filenames ...string) (*Config, error) {
	var (
		cfg = defaultConfig()
		err = configutil.TryLoadConfig(cfg, internal.Name, filenames...)
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
