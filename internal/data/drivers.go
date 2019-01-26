package data

import (
	"github.com/spf13/viper"
	"github.com/stevenxie/api/pkg/data/mongo"
	ess "github.com/unixpickle/essentials"
)

// DriverSet are a set of drivers for storing and retrieving data.
type DriverSet struct {
	Mongo *mongo.DB
}

// LoadDrivers returns a new DriverSet, configured using v.
func LoadDrivers(v *viper.Viper) (*DriverSet, error) {
	mongo, err := mongo.ConnectUsing(v)
	if err != nil {
		return nil, ess.AddCtx("data: connecting to Mongo", err)
	}
	return &DriverSet{Mongo: mongo}, nil
}

// Close closes all active data drivers in the DriverSet.
func (d *DriverSet) Close() error {
	return nil
}
