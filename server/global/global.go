package global

import (
	"gorm.io/gorm"

	"worddictation/server/config"
)

var (
	DB     *gorm.DB
	Config *config.Config
)
