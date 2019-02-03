package client

import (
	"errors"
	"net/http"
	"net/http/cookiejar"

	validator "gopkg.in/validator.v2"

	"github.com/spf13/viper"
)

// Client is capable of retrieving data from the Airtable API.
type Client struct {
	HC  *http.Client
	Jar *cookiejar.Jar

	cfg *Config
}

// New creates a new Airtable client.
func New(cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("client: cannot create client with nil config")
	}
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	// Configure components.
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	hc := &http.Client{Jar: jar}

	// Create client.
	return &Client{
		HC:  hc,
		Jar: jar,
		cfg: cfg,
	}, nil
}

// NewUsing creates a new Airtable client using a config derived from v.
func NewUsing(v *viper.Viper) (*Client, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return New(cfg)
}
