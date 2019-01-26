package mongo

import (
	"context"
	"time"

	defaults "gopkg.in/mcuadros/go-defaults.v1"

	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	ess "github.com/unixpickle/essentials"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "mongo"

// Config describes the settings for connecting to a Postgres database.
type Config struct {
	Host string `ms:"host" default:"localhost"`
	Port string `ms:"port" default:"27017"`
	User string `ms:"user" valid:"nonzero"`
	Pass string `ms:"pass" valid:"nonzero"`
	DB   string `ms:"db"   valid:"nonzero"`

	ConnectTimeout   time.Duration `ms:"connect_timeout"`
	OperationTimeout time.Duration `ms:"operation_timeout"`
}

const defaultConnectTimeout = "10s"

// SetDefaults loads the default values for any zeroed fields in the Config.
func (cfg *Config) SetDefaults() {
	defaults.SetDefaults(cfg)

	if cfg.ConnectTimeout == 0 {
		var err error
		cfg.ConnectTimeout, err = time.ParseDuration(defaultConnectTimeout)
		if err != nil {
			panic(err)
		}
	}
}

// OperationContext returns an operational context, with a timeout as specified
// by cfg.OperationTimeout.
//
// If cfg.OperationTimeout is zero, OperationContext returns a background context
// with a no-op context.CancelFunc.
func (cfg *Config) OperationContext() (context.Context, context.CancelFunc) {
	bg := context.Background()
	if cfg.OperationTimeout == 0 {
		return bg, noop
	}
	return context.WithTimeout(bg, cfg.OperationTimeout)
}

func noop() {}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	v = v.Sub(Namespace)
	if v == nil {
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
