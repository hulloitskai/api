package config

import (
	"time"

	validator "gopkg.in/go-validator/validator.v2"
)

// Config maps to a configuration YAML that can configure programs in this
// package.
type Config struct {
	About struct {
		Gist struct {
			ID   string `yaml:"id"   valid:"nonzero"`
			File string `yaml:"file" valid:"nonzero"`
		} `yaml:"gist"`
	} `yaml:"about"`

	Commits struct {
		Limit        *int           `yaml:"limit"`
		PollInterval *time.Duration `yaml:"pollInterval"`
	} `yaml:"commits"`

	Availability struct {
		GCal struct {
			CalendarIDs []string `yaml:"calendarIDs"`
		} `yaml:"gcal"`
	} `yaml:"availability"`

	NowPlaying struct {
		PollInterval *time.Duration `yaml:"pollInterval"`
	} `yaml:"nowPlaying"`

	Location struct {
		PollInterval *time.Duration `yaml:"pollInterval"`
	} `yaml:"location"`

	// Miscellaneous:
	ShutdownTimeout *time.Duration `yaml:"shutdownTimeout"`
}

func defaultConfig() *Config {
	return new(Config)
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error { return validator.Validate(cfg) }
