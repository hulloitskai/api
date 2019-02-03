package mongo

import (
	"errors"
	"fmt"

	defaults "github.com/mcuadros/go-defaults"
	m "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/spf13/viper"
	ess "github.com/unixpickle/essentials"
	validator "gopkg.in/validator.v2"
)

// A Provider provides various services that use Mongo as an underlying data
// store.
type Provider struct {
	Client *m.Client
	DB     *m.Database

	*MoodService

	cfg *Config
}

// NewProvider returns a new Provider.
func NewProvider(cfg *Config) (*Provider, error) {
	if cfg == nil {
		return nil, errors.New("mongo: config must not be nil")
	}

	defaults.SetDefaults(cfg)
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	client, err := m.NewClient(cfg.Connstr())
	if err != nil {
		return nil, err
	}

	return &Provider{
		Client: client,
		cfg:    cfg,
	}, nil
}

// NewProviderUsing returns a new Provider that is configured using v.
func NewProviderUsing(v *viper.Viper) (*Provider, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return NewProvider(cfg)
}

// Open opens a connection to Mongo, and initializes p.DB as well as other
// internal services.
func (p *Provider) Open() error {
	ctx, cancel := p.cfg.ConnectContext()
	defer cancel()

	if err := p.Client.Connect(ctx); err != nil {
		return err
	}

	// Initialize p.DB.
	db := p.Client.Database(p.cfg.DB)
	if db == nil {
		return fmt.Errorf("mongo: no such database '%s'", p.cfg.DB)
	}
	p.DB = db

	// Initialize services.
	var err error
	p.MoodService, err = newMoodService(p.DB, p.cfg.OperationContext)
	return ess.AddCtx("mongo: creating mood service", err)
}

// Close closes the Provider's underlying Mongo connection.
func (p *Provider) Close() error {
	ctx, cancel := p.cfg.ConnectContext()
	defer cancel()
	return p.Client.Disconnect(ctx)
}
