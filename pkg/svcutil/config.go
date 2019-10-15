package svcutil

import "github.com/sirupsen/logrus"

// WithLogger configures a service to write logs using log.
func WithLogger(log *logrus.Entry) BasicOption {
	return func(cfg *BasicConfig) { cfg.Logger = log }
}

type (
	// A BasicConfig is a basic configuration for a service.
	BasicConfig struct {
		Logger *logrus.Entry
	}

	// A BasicOption modifies a BasicConfig.
	BasicOption func(*BasicConfig)
)
