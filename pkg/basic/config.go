package basic

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.stevenxie.me/gopkg/logutil"
)

// WithLogger configures a basic component to write logs using log.
func WithLogger(log *logrus.Entry) Option {
	return func(opt *Options) { opt.Logger = log }
}

// WithTracer configures a basic component to trace calls with t.
func WithTracer(t opentracing.Tracer) Option {
	return func(opt *Options) { opt.Tracer = t }
}

type (
	// Options configures a basic component.
	Options struct {
		Logger *logrus.Entry
		Tracer opentracing.Tracer
	}

	// An Option modifies an Options.
	Option func(*Options)
)

// DefaultOptions creates an Options with sensible defaults.
func DefaultOptions() Options {
	return Options{
		Logger: logutil.NoopEntry(),
		Tracer: new(opentracing.NoopTracer),
	}
}

// BuildOptions builds an Options from opts.
//
// It uses DefaultOptions as a base.
func BuildOptions(opts ...Option) Options {
	opt := DefaultOptions()
	for _, apply := range opts {
		apply(&opt)
	}
	return opt
}

// ApplyOptions applies opts to opt.
func ApplyOptions(opt *Options, opts ...Option) {
	for _, apply := range opts {
		apply(opt)
	}
}
