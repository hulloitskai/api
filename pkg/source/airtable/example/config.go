package main

import (
	"github.com/caarlos0/env"
)

type config struct {
	apiKey string `env:"API_KEY,required"`
	baseID string `env:"BASE_ID,required"`
}

func (cfg *config) APIKey() string {
	return cfg.apiKey
}

func (cfg *config) BaseID() string {
	return cfg.baseID
}

func readConfig() (*config, error) {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
