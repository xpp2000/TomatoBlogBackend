package dao

import (
	"gorm.io/gorm"
)

type BaseDao struct {
	Orm *gorm.DB
}

func NewBaseDao(db *gorm.DB) *BaseDao {
	return &BaseDao{
		Orm: db,
	}
}
