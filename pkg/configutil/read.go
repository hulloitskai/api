package configutil

import (
	"io"

	"go.stevenxie.me/api/pkg/zero"
	yaml "gopkg.in/yaml.v2"
)

// Read reads in YAML data from an io.Reader into cfg.
func Read(cfg zero.Interface, r io.Reader) error {
	return yaml.NewDecoder(r).Decode(cfg)
}
