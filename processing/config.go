package processing

import (
	defaults "github.com/mcuadros/go-defaults"
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/util"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "jobserver"

// Config describes the settings for connecting to a Postgres database.
type Config struct {
	RedisAddr      string `ms:"redisAddr" default:":6379"`
	FetchMoodsCron string `ms:"fetchMoodsCron" default:"@every 5m"`
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	if v = v.Sub(Namespace); v == nil {
		v = viper.New()
	}
	cfg := new(Config)
	if err := v.Unmarshal(cfg, util.DecoderConfigOption); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SetDefaults sets the default values for cfg.
func (cfg *Config) SetDefaults() {
	defaults.SetDefaults(cfg)
}
