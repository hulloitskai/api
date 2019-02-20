package main

import (
	errors "golang.org/x/xerrors"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/airtable"
	"github.com/stevenxie/api/data/mongo"
)

type provider struct {
	airtable *airtable.Provider
	mongo    *mongo.Provider

	*airtable.MoodSource
	*mongo.MoodService
}

func newProvider(v *viper.Viper) (*provider, error) {
	if v == nil {
		v = viper.New()
	}

	airtable, err := airtable.NewProviderViper(v.Sub(airtable.Namespace))
	if err != nil {
		return nil, errors.Errorf("creating airtable provider: %w", err)
	}
	mongo, err := mongo.NewProviderFromViper(v.Sub(mongo.Namespace))
	if err != nil {
		return nil, errors.Errorf("creating mongo provider: %w", err)
	}

	return &provider{
		airtable:   airtable,
		mongo:      mongo,
		MoodSource: airtable.MoodSource(),
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
