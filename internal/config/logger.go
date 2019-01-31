package config

import (
	"os"

	"go.uber.org/zap"
)

// BuildLogger builds a preconfigured zap.Logger.
func BuildLogger() (*zap.SugaredLogger, error) {
	var (
		raw *zap.Logger
		err error
	)
	if os.Getenv("GOENV") == "development" {
		raw, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewProductionConfig()
		cfg.Encoding = "console"
		raw, err = cfg.Build()
	}
	if err != nil {
		return nil, err
	}
	return raw.Sugar(), nil
}
