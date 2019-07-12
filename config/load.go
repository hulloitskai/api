package config

import (
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/internal/info"
)

// DefaultFilepaths are paths to look for config files.
var DefaultFilepaths = []string{
	info.Namespace + ".yaml",
	"/etc/" + info.Namespace + "/config.yaml",
}

// LoadFile reads a Config from a file.
//
// It also reads in values from the environment.
func LoadFile(name string) (*Config, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "config: opening file")
	}
	defer file.Close()

	cfg, err := ReadFrom(file)
	if err != nil {
		return nil, err
	}

	if err = file.Close(); err != nil {
		return nil, errors.Wrap(err, "config: closing file")
	}
	return cfg, nil
}

// Load attempts to the load a Config by checking for files located at
// DefaultFilepaths.
//
// It also reads in values from the environment.
func Load() (*Config, error) {
	for _, path := range DefaultFilepaths {
		_, err := os.Stat(path)
		if err == nil {
			return LoadFile(path)
		}
		if !os.IsNotExist(err) {
			return nil, errors.Wrapf(err, "config: checking file '%s'", path)
		}
	}
	return nil, ErrNoFilesFound
}

// ErrNoFilesFound is returned when none of the config files located at
// DefaultFilepaths exist.
var ErrNoFilesFound = errors.New("config: no config files were found")

// ReadFrom reads a Config from an io.Reader.
//
// It also reads in values from the environment.
func ReadFrom(r io.Reader) (*Config, error) {
	var (
		dec = yaml.NewDecoder(r)
		cfg = defaultConfig()
	)
	if err := dec.Decode(cfg); err != nil {
		return nil, errors.Wrap(err, "config: decoding YAML")
	}
	return cfg, nil
}
