package airtable

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/common"
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
	if v = v.Sub(Namespace); v == nil {
		v = viper.New()
	}

	v.SetEnvPrefix(Namespace)
	if err := v.BindEnv("api_key"); err != nil {
		return nil, ess.AddCtx("airtable: binding viper envvars", err)
	}
	var (
		cfg = new(Config)
		err = v.Unmarshal(cfg, common.DecoderConfigOption)
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
