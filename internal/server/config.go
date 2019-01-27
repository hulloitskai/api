package server

import (
	"context"
	"time"

	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// Config holds options for configuring a Server.
type Config struct {
	ShutdownTimeout time.Duration `ms:"shutdown_timeout" default:"2s"`
}

// ShutdownContext returns a context with the appropriate timeout for a server
// shutdown.
func (cfg *Config) ShutdownContext() (context.Context, context.CancelFunc) {
	bg := context.Background()
	if cfg.ShutdownTimeout == 0 {
		return bg, noop
	}
	return context.WithTimeout(bg, cfg.ShutdownTimeout)
}

func noop() {}

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
