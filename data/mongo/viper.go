package mongo

import (
	"github.com/stevenxie/api/internal/util"

	"github.com/spf13/viper"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "mongo"

// NewProviderFromViper creates a new Provider, configured using Viper.
func NewProviderFromViper(v *viper.Viper) (*Provider, error) {
	if v == nil {
		v = viper.New()
	}

	util.BindViperEnv(v, Namespace, "uri", "db")
	p, err := NewProvider(v.GetString("uri"), v.GetString("db"))
	if err != nil {
		return nil, err
	}

	util.BindViperEnv(v, Namespace, "connectTimeout", "operationTimeout")
	if v.IsSet("connectTimeout") {
		p.SetConnectTimeout(v.GetDuration("connectTimeout"))
	}
	if v.IsSet("operationTimeout") {
		p.SetOperationTimeout(v.GetDuration("operationTimeout"))
	}

	return p, nil
}
