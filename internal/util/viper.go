package util

import (
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

var (
	// DecoderConfigOption is the common viper.DecoderConfigOption used by
	// api packages.
	DecoderConfigOption = func(dc *ms.DecoderConfig) {
		dc.TagName = "ms"
	}
)

// LoadLocalViper loads a YAML Viper config in the working directory with the
// provided name.
func LoadLocalViper(name string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigName(name)
	v.AddConfigPath(".")
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	return v, nil
}
