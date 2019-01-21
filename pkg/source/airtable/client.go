package airtable

import (
	"errors"

	at "github.com/fabioberger/airtable-go"
	ess "github.com/unixpickle/essentials"
)

type Client struct {
	cfg Config
	c   at.Client
}

func New(config *Config) (*Client, error) {
	if config == nil {
		return nil, errors.New("airtable: nil config")
	}
	cfg := *config
	cfg.setDefaults()

	c, err := at.New(cfg.APIKey, cfg.BaseID)
	if err != nil {
		return nil, ess.AddCtx("airtable", err)
	}

	return &Client{c: *c, cfg: cfg}, nil
}
