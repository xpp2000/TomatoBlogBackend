package global

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Config *AppConfig
	Logger *zap.SugaredLogger
	DB     *gorm.DB
)
