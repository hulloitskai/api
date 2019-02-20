package server

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/util"
)

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "server"

// NewFromViper creates a new Server, configured using Viper.
func NewFromViper(p Provider, v *viper.Viper) *Server {
	if v == nil {
		v = viper.New()
	}
	util.BindViperEnv(v, Namespace, "shutdownTimeout")
	srv := New(p)

	if v.IsSet("shutdownTimeout") {
		srv.SetShutdownTimeout(v.GetDuration("shutdownTimeout"))
	}
	return srv
}
