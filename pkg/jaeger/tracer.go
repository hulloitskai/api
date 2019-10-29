package jaeger

import (
	"io"

	"github.com/cockroachdb/errors"
	"github.com/imdario/mergo"
	"go.stevenxie.me/gopkg/zero"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
)

// NewTracer creates a new jaeger.Tracer.
func NewTracer(name string, opts ...Option) (opentracing.Tracer, io.Closer, error) {
	opt := Options{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	for _, apply := range opts {
		apply(&opt)
	}

	// Construct c from env, and merge cfg with overwrite.
	c, err := config.FromEnv()
	if err != nil {
		return nil, nil, errors.Wrap(
			err,
			"jaeger: configuring using environment variables",
		)
	}
	merges := []struct {
		Name string
		Dst  zero.Interface
		Src  zero.Interface
	}{
		{Name: "Sampler", Dst: c.Sampler, Src: opt.Sampler},
		{Name: "Reporter", Dst: c.Reporter, Src: opt.Reporter},
	}
	for _, m := range merges {
		if m.Dst == nil || m.Src == nil {
			continue
		}
		if err = mergo.Merge(m.Dst, m.Src, mergo.WithOverride); err != nil {
			return nil, nil, errors.Wrapf(err, "jaeger: merging %s", m.Name)
		}
	}

	return c.New(name)
}

// WithSamplerConfig configures a jaeger.Tracer to use the provided
// config.SamplerConfig.
func WithSamplerConfig(cfg *config.SamplerConfig) Option {
	return func(opt *Options) {
		if err := mergo.Merge(opt.Sampler, cfg, mergo.WithOverride); err != nil {
			panic(errors.Wrap(err, "jaeger: failed to merge configs"))
		}
	}
}

// WithReporterConfig configures a jaeger.Tracer to use the provided
// config.ReporterConfig.
func WithReporterConfig(cfg *config.ReporterConfig) Option {
	return func(opt *Options) {
		if err := mergo.Merge(opt.Reporter, cfg, mergo.WithOverride); err != nil {
			panic(errors.Wrap(err, "jaeger: failed to merge configs"))
		}
	}
}

type (
	// An Options is a susbet of a config.Configuration.
	Options struct {
		Sampler  *config.SamplerConfig  `yaml:"sampler"`
		Reporter *config.ReporterConfig `yaml:"reporter"`
	}

	// An Option modifies an Options.
	Option func(*Options)
)
