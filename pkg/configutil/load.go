package configutil

import (
	stderrs "errors"
	"os"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/api/pkg/zero"
)

// LoadFile loads a Config from a file.
func LoadFile(cfg zero.Interface, filepath string) error {
	// Open file.
	file, err := os.Open(filepath)
	if err != nil {
		return errors.Wrap(err, "configutil: opening file")
	}
	defer file.Close()

	// Read file.
	if err = Read(cfg, file); err != nil {
		return err
	}

	// Close file.
	if err = file.Close(); err != nil {
		return errors.Wrap(err, "configutil: closing file")
	}
	return nil
}

// TryLoadFiles tries to load a Config from multiple possible filepaths; it
// loads a Config from sthe first valid file that it finds.
//
// It also reads in values from the environment, which are overridden by
// values from the file.
func TryLoadFiles(cfg zero.Interface, filepaths ...string) error {
	for _, path := range filepaths {
		_, err := os.Stat(path)
		if err == nil {
			return LoadFile(cfg, path)
		}
		if !os.IsNotExist(err) {
			return errors.Wrapf(err, "configutil: checking file '%s'", path)
		}
	}
	return ErrNotFound
}

// ErrNotFound is returned by TryLoadFiles and Load if no config files were
// found at any of the possible filepaths.
var ErrNotFound = stderrs.New("configutil: no config files were found at " +
	"the specified paths")
