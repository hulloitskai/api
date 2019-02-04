package airtable

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/airtable"
	ess "github.com/unixpickle/essentials"
)

// A Provider provides data using Airtable as an underlying datasource.
type Provider struct {
	*Config
	*airtable.Client

	*MoodSource
}

// New returns a new Provider.
func New(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, errors.New("airtable: cannot create provider with nil config")
	}
	cfg.SetDefaults()

	c, err := airtable.New(cfg.ClientConfig)
	if err != nil {
		return nil, ess.AddCtx("airtable: creating client", err)
	}

	return &Provider{
		Config:     cfg,
		Client:     c,
		MoodSource: newMoodSource(c, cfg.MoodSourceConfig),
	}, nil
}

// NewUsing returns a Provider that is configured using cfg.
func NewUsing(v *viper.Viper) (*Provider, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}
