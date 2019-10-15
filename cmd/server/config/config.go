package config

import (
	"time"

	"github.com/cockroachdb/errors"
	"go.stevenxie.me/api/auth/airtable"

	validation "github.com/go-ozzo/ozzo-validation"
)

// Config maps to a configuration YAML that can configure programs in this
// package.
type Config struct {
	ShutdownTimeout *time.Duration `yaml:"shutdownTimeout"`

	About struct {
		Gist struct {
			ID   string `yaml:"id"`
			File string `yaml:"file"`
		} `yaml:"gist"`
	} `yaml:"about"`

	Git struct {
		Precacher struct {
			Enabled  bool          `yaml:"enabled"`
			Interval time.Duration `yaml:"interval"`
			Limit    *int          `yaml:"limit"`
		} `yaml:"precacher"`
	} `yaml:"git"`

	Scheduling struct {
		GCal struct {
			CalendarIDs []string `yaml:"calendarIDs"`
		} `yaml:"gcal"`
	} `yaml:"scheduling"`

	Music struct {
		Streamer struct {
			PollInterval time.Duration `yaml:"pollInterval"`
		} `yaml:"streamer"`
	} `yaml:"music"`

	Location struct {
		Precacher struct {
			Enabled  bool          `yaml:"enabled"`
			Interval time.Duration `yaml:"interval"`
		} `yaml:"precacher"`

		Here struct {
			AppID string `yaml:"appID"`
		} `yaml:"here"`

		CurrentRegion struct {
			GeocodeLevel string `yaml:"geocodeLevel"`
		} `yaml:"currentRegion"`
	}

	Auth struct {
		Airtable struct {
			Selector      airtable.PermsSelector `yaml:"selector"`
			AccessRecords struct {
				Enabled  bool                    `yaml:"enabled"`
				Selector airtable.AccessSelector `yaml:"selector"`
			} `yaml:"accessRecords"`
		} `yaml:"airtable"`
	} `yaml:"auth"`
}

func defaultConfig() *Config {
	cfg := new(Config)

	// Default location precacher settings.
	{
		cfg := &cfg.Location.Precacher
		cfg.Enabled = true
		cfg.Interval = 2 * time.Minute
	}

	// Default Git precacher settings.
	{
		cfg := &cfg.Git.Precacher
		cfg.Enabled = true
		cfg.Interval = 10 * time.Minute
	}

	// Default music streamer settings.
	{
		cfg := &cfg.Music.Streamer
		cfg.PollInterval = time.Second
	}

	// Default Airtable settings.
	{
		cfg := &cfg.Auth.Airtable
		cfg.Selector = airtable.DefaultPermsSelector()
		cfg.AccessRecords.Selector = airtable.DefaultAccessSelector()
	}

	return cfg
}

// Validate returns an error if the Config is not valid.
func (cfg *Config) Validate() error {
	{
		gist := &cfg.About.Gist
		if err := validation.ValidateStruct(
			gist,
			validation.Field(&gist.ID, validation.Required),
			validation.Field(&gist.File, validation.Required),
		); err != nil {
			return errors.Wrap(err, "validating About.Gist")
		}
	}

	{
		location := cfg.Location
		if err := validation.Validate(
			location.Here.AppID,
			validation.Required,
		); err != nil {
			return errors.Wrap(err, "validating Location.Here.AppID")
		}

		if err := validation.Validate(
			location.CurrentRegion.GeocodeLevel,
			validation.Required,
		); err != nil {
			return errors.Wrap(
				err,
				"validating Location.CurrentRegion.GeocodeLevel",
			)
		}

		if err := validation.Validate(
			location.Precacher.Interval,
			validation.Min(1),
		); err != nil {
			return errors.Wrap(err, "validating Location.Precacher.Interval")
		}
	}

	if err := validation.Validate(
		cfg.Scheduling.GCal.CalendarIDs,
		validation.Required,
	); err != nil {
		return errors.Wrap(err, "validating Scheduling.GCal.CalendarIDs")
	}

	{
		at := &cfg.Auth.Airtable
		if err := validation.Validate(&at.Selector); err != nil {
			return errors.Wrap(err, "validating Auth.Airtable.Selector")
		}
		if err := validation.Validate(&at.AccessRecords.Selector); err != nil {
			return errors.Wrap(err, "validating Auth.Airtable.Access.Selector")
		}
	}

	return nil
}
