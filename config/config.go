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
		Polling struct {
			Enabled  bool          `yaml:"enabled"`
			Limit    int           `yaml:"limit"`
			Interval time.Duration `yaml:"interval"`
		} `yaml:"polling"`
	} `yaml:"commits"`

	Availability struct {
		GCal struct {
			CalendarIDs []string `yaml:"calendarIDs"`
		} `yaml:"gcal"`
	} `yaml:"availability"`

	Music struct {
		Polling struct {
			Interval time.Duration `yaml:"interval"`
		} `yaml:"polling"`
	} `yaml:"music"`

	Location struct {
		Polling struct {
			Enabled  bool          `yaml:"enabled"`
			Interval time.Duration `yaml:"interval"`
		} `yaml:"polling"`

		Here struct {
			AppID string `yaml:"appID" valid:"nonzero"`
		} `yaml:"here"`

		Airtable struct {
			BaseID string `yaml:"baseID"`
			Table  string `yaml:"table"`
			View   string `yaml:"view"`
		} `yaml:"airtable"`
	} `yaml:"location"`

	// Miscellaneous:
	ShutdownTimeout *time.Duration `yaml:"shutdownTimeout"`
}

func defaultConfig() *Config {
	cfg := new(Config)

	cfg.Commits.Polling.Limit = 5
	cfg.Commits.Polling.Interval = time.Minute

	cfg.Music.Polling.Interval = 5 * time.Second
	cfg.Location.Polling.Interval = time.Minute
	return cfg
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error { return validator.Validate(cfg) }
