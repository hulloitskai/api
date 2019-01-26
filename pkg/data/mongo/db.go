package mongo

import (
	m "github.com/mongodb/mongo-go-driver/mongo"
)

// A DB holds a mongo.Database and a Config.
type DB struct {
	*m.Database
	Config *Config
}
