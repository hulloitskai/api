package config

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/stevenxie/api/internal/info"
)

// BuildViper returns a viper.Viper instance that is set to read a YAML config
// from the host filesystem.
func BuildViper() *viper.Viper {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Configure filepaths to locate configs.
	v.AddConfigPath(".")
	addViperConfigPath(v, "/etc/%s", info.Namespace)
	addViperConfigPath(v, "$HOME/.%s", info.Namespace)
	return v
}

// LoadViper loads a viper.Viper instance from a YAML config on the host
// filesystem.
func LoadViper() (*viper.Viper, error) {
	v := BuildViper()
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}

func addViperConfigPath(v *viper.Viper, format string, a ...interface{}) {
	v.AddConfigPath(fmt.Sprintf(format, a...))
}
