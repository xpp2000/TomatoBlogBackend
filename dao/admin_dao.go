package dao

import (
	"tomatoBlogDB/model"
)

type IAdminDao interface {
	GetAdminByName(name string) (*model.Admin, error)
	CreateAdmin(admin *model.Admin) error
	CheckAdminNameExist(name string) bool
}

type AdminDao struct {
	*BaseDao
}

// 确保实现接口
var _ IAdminDao = (*AdminDao)(nil)

// NewAdminDao 纯净的构造函数，只负责赋值，不搞单例
func NewAdminDao(baseDao *BaseDao) *AdminDao {
	return &AdminDao{
		BaseDao: baseDao, // 老老实实使用外部传进来的依赖
	}
}

// ... 下面的 GetAdminByName, CreateAdmin, CheckAdminNameExist 保持不变 ...
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
