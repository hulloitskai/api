package airtable

import (
	"net/http"
	"net/http/cookiejar"

	validator "gopkg.in/validator.v2"

	"github.com/spf13/viper"
	defaults "gopkg.in/mcuadros/go-defaults.v1"
)

// Client is capable of retrieving data from the Airtable API.
type Client struct {
	HC  *http.Client
	Jar *cookiejar.Jar

	cfg *Config
}

// New creates a new Airtable client.
func New(cfg *Config) *Client {
	defaults.SetDefaults(cfg)
	validator.Validate(cfg)

	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}

	hc := &http.Client{Jar: jar}
	return &Client{
		HC:  hc,
		Jar: jar,
		cfg: cfg,
	}
}

// NewUsing creates a new Airtable client using a config derived from v.
func NewUsing(v *viper.Viper) (*Client, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return New(cfg), nil
}
