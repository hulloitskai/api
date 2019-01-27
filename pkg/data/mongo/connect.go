package mongo

import (
	"context"
	"fmt"

	validator "gopkg.in/validator.v2"

	defaults "github.com/mcuadros/go-defaults"
	m "github.com/mongodb/mongo-go-driver/mongo"
	"github.com/spf13/viper"
	ess "github.com/unixpickle/essentials"
)

// Connect connects to a Mongo database, configured using cfg.
func Connect(cfg *Config) (*DB, error) {
	defaults.SetDefaults(cfg)
	if err := validator.Validate(cfg); err != nil {
		return nil, err
	}

	connstr := fmt.Sprintf(
		"mongodb://%s:%s@%s:%s/%s",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.DB,
	)

	// Configure context with a timeout.
	fmt.Println(cfg.ConnectTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	// Connect to the DB.
	mc, err := m.Connect(ctx, connstr)
	if err != nil {
		return nil, err
	}

	// Ping the DB to validate the connection.
	if err = mc.Ping(ctx, nil); err != nil {
		return nil, ess.AddCtx("mongo: validating connection", err)
	}

	db := mc.Database(cfg.DB)
	if db == nil {
		return nil, fmt.Errorf("mongo: no such database '%s'", cfg.DB)
	}
	return &DB{
		Database: db,
		Config:   cfg,
	}, nil
}

// ConnectUsing connects to a Mongo database, configured using v.
func ConnectUsing(v *viper.Viper) (*DB, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return Connect(cfg)
}
