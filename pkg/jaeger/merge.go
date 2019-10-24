package jaeger

import (
	"github.com/imdario/mergo"
)

// MergeConfigs overwrites fields in dst with non-zero fields from src.
func MergeConfigs(dst *Config, src *Config) error {
	return mergo.Merge(dst, src, mergo.WithOverride)
}
