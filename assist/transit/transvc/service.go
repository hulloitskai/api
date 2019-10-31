package transvc

import (
	"time"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"

	"go.stevenxie.me/api/v2/assist/transit"
	"go.stevenxie.me/gopkg/logutil"
)

// NewService creates a new transit.Service.
func NewService(
	loc transit.LocatorService,
	opts ...ServiceOption,
) transit.Service {
	opt := ServiceOptions{
		Logger:                  logutil.NoopEntry(),
		Tracer:                  new(opentracing.NoopTracer),
		RealtimeSources:         make(map[string]transit.RealtimeSource),
		MaxRealtimeDepartureGap: 3 * time.Hour,
	}
	for _, apply := range opts {
		apply(&opt)
	}
	return &service{
		loc: loc,
		rts: opt.RealtimeSources,

		maxRTDepGap: opt.MaxRealtimeDepartureGap,

		log:    logutil.WithComponent(opt.Logger, (*service)(nil)),
		tracer: opt.Tracer,
	}
}

// WithLogger configures a transit.Service to write logs with log.
func WithLogger(log *logrus.Entry) ServiceOption {
	return func(opt *ServiceOptions) { opt.Logger = log }
}

// WithTracer configures a transit.Service to trace calls with t.
func WithTracer(t opentracing.Tracer) ServiceOption {
	return func(opt *ServiceOptions) { opt.Tracer = t }
}

// WithRealtimeSource configures a transit.Service to use transit.RealtimeSource
// to get realtime departure data for the operators specified by opCodes.
func WithRealtimeSource(
	src transit.RealtimeSource,
	opCodes ...string,
) ServiceOption {
	return func(opt *ServiceOptions) {
		for _, code := range opCodes {
			opt.RealtimeSources[code] = src
		}
	}
}

type (
	// A ServiceOptions configures a transit.Service.
	ServiceOptions struct {
		Logger *logrus.Entry
		Tracer opentracing.Tracer

		// A map of operator codes to real-time data sources.
		RealtimeSources map[string]transit.RealtimeSource

		// The largest departure time for which real-time data will be requested
		// for.
		MaxRealtimeDepartureGap time.Duration
	}

	// A ServiceOption modifies a ServiceOptions.
	ServiceOption func(*ServiceOptions)
)

type service struct {
	loc transit.LocatorService
	rts map[string]transit.RealtimeSource // map of op codes to sources

	maxRTDepGap time.Duration

	log    *logrus.Entry
	tracer opentracing.Tracer
}

var _ transit.Service = (*service)(nil)
