package airtable

import (
	"net/http"
	"net/http/cookiejar"

	defaults "gopkg.in/mcuadros/go-defaults.v1"
)

// Client is capable of retrieving data from the Airtable API.
type Client struct {
	HC  *http.Client
	Jar *cookiejar.Jar

	cfg *Config
}

// New creates a new Airtbale client.
func New(cfg *Config) *Client {
	defaults.SetDefaults(cfg)

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
