package dao

import (
	"sync"
	"tomatoBlogDB/model"
)

var (
	adminDao  *AdminDao
	adminOnce sync.Once
)

type IAdminDao interface {
	GetAdminByName(name string) (*model.Admin, error)
	CreateAdmin(user *model.Admin) error
	CheckAdminNameExist(name string) bool
}

// 2.
type AdminDao struct {
	*BaseDao
}

// check implementation
var _ IAdminDao = (*AdminDao)(nil)

func NewAdminDao() *AdminDao {
	adminOnce.Do(func() {
		adminDao = &AdminDao{
			BaseDao: NewBaseDao(),
		}
	})
	return adminDao
}

func (d *AdminDao) GetAdminByName(name string) (*model.Admin, error) {
	var admin model.Admin
	err := d.Orm.Where("name=?", name).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

func (d *AdminDao) CreateAdmin(admin *model.Admin) error {
	return d.Orm.Create(admin).Error
}

func (d *AdminDao) CheckAdminNameExist(name string) bool {
	var count int64
	d.Orm.Model(&model.Admin{}).Where("name=?", name).Count(&count)
	return count > 0
}
