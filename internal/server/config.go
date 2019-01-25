package server

import (
	"time"

	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config holds options for configuring a Server.
type Config struct {
	ShutdownTimeout time.Duration `ms:"shutdown_timeout"`
}

// SetDefaults sets zeroed values in Config to sensible default values.
func (c *Config) SetDefaults() error {
	var err error
	if c.ShutdownTimeout == 0 {
		c.ShutdownTimeout, err = time.ParseDuration("2s")
	}
	return err
}

// ConfigFromViper reads a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	v = v.Sub("server")
	if v == nil {
		return new(Config), nil
	}

	var (
		cfg Config
		err = v.Unmarshal(&cfg, func(dc *ms.DecoderConfig) {
			dc.TagName = "ms"
			dc.DecodeHook = ms.ComposeDecodeHookFunc(dc.DecodeHook,
				ms.StringToTimeDurationHookFunc)
		})
	)
	return &cfg, err
}
