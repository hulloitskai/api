package routes

import (
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

// A Config is used to configure a Router.
type Config struct {
	DB     *gorm.DB
	Logger *zap.SugaredLogger
}
