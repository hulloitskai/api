package jobserver

import "github.com/spf13/viper"

// Namespace is the package namespace, used for configuring environment
// variables, etc.
const Namespace = "jobserver"

// NewFromViper creates a new Server, configured using Viper.
func NewFromViper(p Provider, v *viper.Viper) *Server {
	if v == nil {
		v = viper.New()
	}
	srv := New(p, v.GetString("redisAddr"))

	if v.IsSet("fetchMoodsCron") {
		srv.SetFetchMoodsCron(v.GetString("fetchMoodsCron"))
	}
	return srv
}
