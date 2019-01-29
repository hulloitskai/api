package postgres

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	ess "github.com/unixpickle/essentials"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "postgres"

// Config describes the settings for connecting to a Postgres database.
type Config struct {
	Host    string `ms:"host"    default:"localhost"`
	Port    string `ms:"post"    default:"5432"`
	User    string `ms:"user"    default:"postgres"`
	Pass    string `ms:"pass"    valid:"nonzero"`
	DB      string `ms:"db"      valid:"nonzero"`
	SSLMode string `ms:"sslmode" default:"disable"`
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	if v = v.Sub(Namespace); v == nil {
		v = viper.New()
	}

	v.SetEnvPrefix(Namespace)
	if err := v.BindEnv("pass"); err != nil {
		return nil, ess.AddCtx("postgres: binding viper envvars", err)
	}
	var (
		cfg Config
		err = v.Unmarshal(&cfg, func(dc *ms.DecoderConfig) {
			dc.TagName = "ms"
		})
	)
	return &cfg, err
}
