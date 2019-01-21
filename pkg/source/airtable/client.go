package airtable

import (
	"errors"

	at "github.com/fabioberger/airtable-go"
	ess "github.com/unixpickle/essentials"
)

type Client struct {
	c at.Client
}

func New(cfg Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("airtable: nil config")
	}

	c, err := at.New(cfg.APIKey(), cfg.BaseID())
	if err != nil {
		return nil, ess.AddCtx("airtable", err)
	}

	return &Client{
		c: *c,
	}, nil
}
