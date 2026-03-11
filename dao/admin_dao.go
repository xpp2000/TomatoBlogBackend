package dao

import (
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"
)

type IAdminDao interface {
	GetAdminByName(name string) (*model.Admin, error)
	GetAdminByID(id uint64) (*model.Admin, error)
	CreateAdmin(admin *model.Admin) error
	CheckAdminNameExist(name string) bool
	UpdateAdminStatus(id uint64, status int) error
	GetAdminList(req dto.AdminListReq) ([]*model.Admin, int64, error)
	DeleteAdmin(id uint64) error
	CheckUnique(field string, value string) (bool, error)
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

func (d *AdminDao) GetAdminByID(id uint64) (*model.Admin, error) {
	var admin model.Admin

	// Preload("Author")
	// 它会在底层自动执行两条 SQL：
	// 1. SELECT * FROM admins WHERE id = ? AND deleted_at IS NULL
	// 2. SELECT * FROM authors WHERE admin_id = ? AND deleted_at IS NULL
	err := d.Orm.Preload("Author").First(&admin, id).Error

	// 严谨的错误处理：如果没查到记录 (gorm.ErrRecordNotFound) 或数据库报错，直接返回 nil
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

// UpdateAdminStatus 修改账号状态
func (d *AdminDao) UpdateAdminStatus(id uint64, status int) error {
	// 使用 Update Column 只更新特定字段，避免触发其他字段的钩子
	return d.Orm.Model(&model.Admin{}).Where("id = ?", id).UpdateColumn("status", status).Error
}

// DeleteAdmin 软删除账号及其关联档案
func (d *AdminDao) DeleteAdmin(id uint64) error {
	// Select("Author") 告诉 GORM，删除 Admin 的同时，把 1:1 关联的 Author 也一起软删除！
	admin := &model.Admin{BaseModel: model.BaseModel{ID: id}}
	return d.Orm.Select("Author").Delete(admin).Error
}

// List
func (d *AdminDao) GetAdminList(req dto.AdminListReq) ([]*model.Admin, int64, error) {
	var admins []*model.Admin
	var total int64

	// 1. init
	db := d.Orm.Model(&model.Admin{})

	// 2. prepare query conditions
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		db = db.Where("name LIKE ? OR mobile LIKE ?", keyword, keyword)
	}
	if req.Role != nil {
		db = db.Where("role = ?", *req.Role)
	}
	if req.Status != nil {
		db = db.Where("status = ?", *req.Status)
	}

	// 3. count candidates
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 4. paginate and order
	offset := (req.Page - 1) * req.PageSize
	err := db.Preload("Author").
		Order("id ASC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&admins).Error

	return admins, total, err
}

// CheckUnique 检查某个字段的值是否已存在
// 参数 field: 数据库字段名 (注意：因为是我们自己在 Service 层写死的，所以没有 SQL 注入风险)
// 参数 value: 要检查的值
func (d *AdminDao) CheckUnique(field string, value string) (bool, error) {
	var count int64

	// 动态拼接 WHERE 条件进行查询
	err := d.Orm.Model(&model.Admin{}).Where(field+" = ?", value).Count(&count).Error

	// 如果 count 大于 0，说明数据已经存在，返回 true
	return count > 0, err
}
