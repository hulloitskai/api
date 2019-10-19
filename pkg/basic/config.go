package basic

import (
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
)

// WithLogger configures a basic component to write logs using log.
func WithLogger(log *logrus.Entry) Option {
	return func(cfg *Config) { cfg.Logger = log }
}

type (
	// A Config configures a basic component.
	Config struct {
		Logger *logrus.Entry
	}

	// A Option modifies a BasicConfig.
	Option func(*Config)
)

// DefaultConfig creates a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		Logger: logutil.NoopEntry(),
	}
}

// BuildConfig builds a Config from a set of Options.
//
// It uses DefaultConfig as a base.
func BuildConfig(opts ...Option) Config {
	cfg := DefaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

// Configure configures a Config with a set of Options.
func Configure(cfg *Config, opts ...Option) {
	for _, opt := range opts {
		opt(cfg)
	}
}
