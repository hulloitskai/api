package transvc

import (
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/assist/transit"
	"go.stevenxie.me/gopkg/logutil"
)

// NewService creates a new transit.Service.
func NewService(
	loc transit.LocatorService,
	opts ...ServiceOption,
) transit.Service {
	cfg := ServiceConfig{
		Logger:          logutil.NoopEntry(),
		RealtimeSources: make(map[string]transit.RealtimeSource),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &service{
		loc: loc,
		rts: cfg.RealtimeSources,
		log: logutil.AddComponent(cfg.Logger, (*service)(nil)),
	}
}

// WithLogger configures a transit.Service to write logs with log.
func WithLogger(log *logrus.Entry) ServiceOption {
	return func(cfg *ServiceConfig) { cfg.Logger = log }
}

// WithRealtimeSource configures a transit.Service to use transit.RealtimeSource
// to get realtime departure data for the operators specified by opCodes.
func WithRealtimeSource(
	src transit.RealtimeSource,
	opCodes ...string,
) ServiceOption {
	return func(cfg *ServiceConfig) {
		for _, code := range opCodes {
			cfg.RealtimeSources[code] = src
		}
	}
}

type (
	service struct {
		loc transit.LocatorService
		rts map[string]transit.RealtimeSource // map of op codes to sources
		log *logrus.Entry
	}

	// A ServiceConfig configures a transit.Service.
	ServiceConfig struct {
		Logger *logrus.Entry

		// RealtimeSources is a map of operator codes to real-time data sources.
		RealtimeSources map[string]transit.RealtimeSource
	}

	// A ServiceOption modifies a ServiceConfig.
	ServiceOption func(*ServiceConfig)
)

var _ transit.Service = (*service)(nil)
