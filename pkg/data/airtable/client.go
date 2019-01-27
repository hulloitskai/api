package airtable

import (
	"net/http"
	"net/http/cookiejar"

	validator "gopkg.in/validator.v2"

	defaults "github.com/mcuadros/go-defaults"
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
	defaults.SetDefaults(cfg)
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	hc := &http.Client{Jar: jar}
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
