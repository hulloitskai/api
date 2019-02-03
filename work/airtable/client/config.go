package client

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/util"
	ess "github.com/unixpickle/essentials"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "airtable"

// Config describes the options for configuring an airtable.Client.
type Config struct {
	APIKey string `ms:"apiKey" validate:"nonzero"`
	BaseID string `ms:"baseId" validate:"nonzero"`
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	if v = v.Sub(Namespace); v == nil {
		v = viper.New()
	}
	if err := v.BindEnv("apiKey", "AIRTABLE_API_KEY"); err != nil {
		panic(ess.AddCtx("airtable: binding viper envvars", err))
	}
	var (
		cfg = new(Config)
		err = v.Unmarshal(cfg, util.DecoderConfigOption)
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
