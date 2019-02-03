package airtable

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/work/airtable/client"
	ess "github.com/unixpickle/essentials"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "airtable"

// A Config is used to configure a Provider.
type Config struct {
	ClientConfig *client.Config
	*MoodSourceConfig
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	ccfg, err := client.ConfigFromViper(v)
	if err != nil {
		return nil, ess.AddCtx("airtable: creating client config", err)
	}

	if v = v.Sub(Namespace); v == nil {
		v = viper.New()
	}
	mscfg, err := MoodSourceConfigFromViper(v)
	if err != nil {
		return nil, err
	}

	return &Config{
		ClientConfig:     ccfg,
		MoodSourceConfig: mscfg,
	}, nil
}

// SetDefaults sets the default values for cfg.
func (cfg *Config) SetDefaults() {
	cfg.MoodSourceConfig.SetDefaults()
}
