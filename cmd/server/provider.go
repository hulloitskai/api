package main

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/mongo"
	errors "golang.org/x/xerrors"
)

// A provider is an implementation of an api.ServiceProvider.
type provider struct {
	mongo *mongo.Provider

	*mongo.MoodService
}

// newProvider creates a new provider, configured using v.
func newProvider(v *viper.Viper) (*provider, error) {
	if v == nil {
		v = viper.New()
	}
	mongo, err := mongo.NewProviderFromViper(v.Sub(mongo.Namespace))
	if err != nil {
		return nil, err
	}
	return &provider{
		mongo: mongo,
	}, nil
}

// Open opens underlying connections to external data sources.
func (p *provider) Open() error {
	if err := p.mongo.Open(); err != nil {
		return errors.Errorf("server: opening mongo.Provider: %w", err)
	}
	p.MoodService = p.mongo.MoodService()
	return nil
}

// Close closes all of the Provider's underlying connections to external data
// sources.
func (p *provider) Close() error {
	return p.mongo.Close()
}
