package postgres

import (
	"fmt"

	validator "gopkg.in/validator.v2"

	defaults "gopkg.in/mcuadros/go-defaults.v1"

	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"

	_ "github.com/jinzhu/gorm/dialects/postgres" // install PG dialect
)

// Open opens a connection to a Postgres DB, configured using cfg.
func Open(cfg *Config) (*gorm.DB, error) {
	defaults.SetDefaults(cfg)
	validator.Validate(cfg)
	connstr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Pass, cfg.DB, cfg.SSLMode,
	)
	return gorm.Open("postgres", connstr)
}

// OpenUsing opens a connection to Postgres using a configuration from
// viper.Viper.
func OpenUsing(v *viper.Viper) (*gorm.DB, error) {
	cfg, err := ConfigFromViper(v)
	if err != nil {
		return nil, err
	}
	return Open(cfg)
}
