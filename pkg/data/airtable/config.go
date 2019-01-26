package airtable

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	ess "github.com/unixpickle/essentials"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "airtable"

// Config describes the options for configuring an airtable.Client.
type Config struct {
	APIKey string `ms:"api_key" validate:"nonzero"`
	BaseID string `ms:"base_id" validate:"nonzero"`
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	v = v.Sub(Namespace)
	if v == nil {
		v = viper.New()
	}

	v.SetEnvPrefix(Namespace)
	if err := v.BindEnv("api_key"); err != nil {
		return nil, ess.AddCtx("airtable: binding viper envvars", err)
	}
	var (
		cfg Config
		err = v.Unmarshal(&cfg, func(dc *ms.DecoderConfig) {
			dc.TagName = "ms"
		})
	)
	return &cfg, err
}
