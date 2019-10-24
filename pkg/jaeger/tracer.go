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
	cfg := Config{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	for _, opt := range opts {
		opt(&cfg)
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
		{Name: "Sampler", Dst: c.Sampler, Src: cfg.Sampler},
		{Name: "Reporter", Dst: c.Reporter, Src: cfg.Reporter},
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

type (
	// A Config is a susbet of a config.Configuration.
	Config struct {
		Sampler  *config.SamplerConfig  `yaml:"sampler"`
		Reporter *config.ReporterConfig `yaml:"reporter"`
	}

	// A Option modifies a SimpleConfig.
	Option func(*Config)
)
