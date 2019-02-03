package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/util"
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

	ConnectTimeout   time.Duration `ms:"connect_timeout" default:"20s"`
	OperationTimeout time.Duration `ms:"operation_timeout"`
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
		cfg = new(Config)
		err = v.Unmarshal(cfg, util.DecoderConfigOption)
	)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// OperationContext returns an operational context, with a timeout as specified
// by cfg.OperationTimeout.
//
// If cfg.OperationTimeout is zero, OperationContext returns a background
// context with a no-op context.CancelFunc.
func (cfg *Config) OperationContext() (context.Context, context.CancelFunc) {
	bg := context.Background()
	if cfg.OperationTimeout == 0 {
		return bg, util.Noop
	}
	return context.WithTimeout(bg, cfg.OperationTimeout)
}

// ConnectContext returns a context with a timeout specified by
// cfg.ConnectTimeout.
func (cfg *Config) ConnectContext() (context.Context, context.CancelFunc) {
	bg := context.Background()
	if cfg.ConnectTimeout == 0 {
		return bg, util.Noop
	}
	return context.WithTimeout(bg, cfg.ConnectTimeout)
}

// Connstr builds a Mongo connection string.
func (cfg *Config) Connstr() string {
	return fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB,
	)
}
