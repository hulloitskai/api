package util

import (
	"fmt"
	"strings"

	errors "golang.org/x/xerrors"

	"github.com/iancoleman/strcase"
	ms "github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

// DecodeShortTag is a viper.DecoderConfigOption that uses the short tag 'ms',
// rather than 'mapstructure'.
func DecodeShortTag(dc *ms.DecoderConfig) { dc.TagName = "ms" }

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

// BindViperEnv binds all keys to env variables in Viper. Each key is prefixed
// with prefix, screaming-snake-cased, and uppercased.
//
// For example, the Viper key 'decodeJSONlimit' would be bound to the
// environment key '${PREFIX}_DECODE_JSON_LIMIT'.
func BindViperEnv(v *viper.Viper, prefix string, keys ...string) {
	prefix = strings.ToUpper(prefix)
	for _, key := range keys {
		if err := v.BindEnv(
			key,
			fmt.Sprintf("%s_%s", prefix, strcase.ToScreamingSnake(key)),
		); err != nil {
			panic(errors.Errorf("util: binding Viper to env keys: %w", err))
		}
	}
}
