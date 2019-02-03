package data

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/data/mongo"
)

// Config describes the settings for connecting to a Postgres database.
type Config struct {
	MongoConfig *mongo.Config
}

// ConfigFromViper parses a Config from a viper.Viper instance.
func ConfigFromViper(v *viper.Viper) (*Config, error) {
	mcfg, err := mongo.ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return &Config{
		MongoConfig: mcfg,
	}, nil
}
