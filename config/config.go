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
		Limit        int           `yaml:"limit"`
		PollInterval time.Duration `yaml:"pollInterval"`
	} `yaml:"commits"`

	Availability struct {
		GCal struct {
			CalendarIDs []string `yaml:"calendarIDs"`
		} `yaml:"gcal"`
	} `yaml:"availability"`

	Music struct {
		PollInterval time.Duration `yaml:"pollInterval"`
	} `yaml:"music"`

	Location struct {
		PollInterval time.Duration `yaml:"pollInterval"`
		Here         struct {
			AppID string `yaml:"appID"`
		} `yaml:"here"`
	} `yaml:"location"`

	// Miscellaneous:
	ShutdownTimeout *time.Duration `yaml:"shutdownTimeout"`
}

func defaultConfig() *Config {
	cfg := new(Config)
	cfg.Commits.Limit = 5
	cfg.Commits.PollInterval = time.Minute
	cfg.Music.PollInterval = 5 * time.Second
	cfg.Location.PollInterval = time.Minute
	return cfg
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error { return validator.Validate(cfg) }
