package data

import (
	"errors"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/mongo"
)

// A Provider is an implementation of an api.ServiceProvider.
type Provider struct {
	Mongo *mongo.Provider

	*mongo.MoodService
}

// NewProvider returns a new Provider, configured using cfg.
func NewProvider(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, errors.New("provider: config must not be nil")
	}

	mp, err := mongo.NewProvider(cfg.MongoConfig)
	if err != nil {
		return nil, err
	}

	return &Provider{Mongo: mp}, nil
}

// NewProviderUsing returns a new Provider, configured using v.
func NewProviderUsing(v *viper.Viper) (*Provider, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return NewProvider(cfg)
}

// Open opens underlying connections to external data sources.
func (p *Provider) Open() error {
	if err := p.Mongo.Open(); err != nil {
		return err
	}
	p.MoodService = p.Mongo.MoodService
	return nil
}

// Close closes all of the Provider's underlying connections to external data
// sources.
func (p *Provider) Close() error {
	return p.Mongo.Close()
}
