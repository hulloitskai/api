package airtable

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/util"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "airtable"

// NewClientViper creates a new Client, configured using Viper.
func NewClientViper(v *viper.Viper) (*Client, error) {
	if v == nil {
		v = viper.New()
	}
	util.BindViperEnv(v, Namespace, "apiKey", "baseId")
	return NewClient(v.GetString("apiKey"), v.GetString("baseId"))
}

// NewProviderViper creates a new Provider, configured using Viper.
func NewProviderViper(v *viper.Viper) (*Provider, error) {
	if v == nil {
		v = viper.New()
	}

	util.BindViperEnv(v, Namespace, "apiKey", "baseId")
	p, err := NewProvider(v.GetString("apiKey"), v.GetString("baseId"))
	if err != nil {
		return nil, err
	}

	if v := v.Sub("moodSource"); v != nil {
		if limit := v.GetInt("fetchLimit"); limit > 0 {
			p.MoodSource().SetFetchLimit(limit)
		}
	}

	return p, nil
}
