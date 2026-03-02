package model

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Admin struct {
	BaseModel
	// 核心信息
	Name     string `json:"name" gorm:"type:varchar(64);not null;comment:user name"`
	RealName string `json:"real_name,omitempty" gorm:"type:varchar(128);comment:real name"`

	// 联系方式 (添加唯一索引)
	Mobile string `json:"mobile" gorm:"type:char(11);uniqueIndex;not null;comment:+86 phone number"`
	Email  string `json:"email,omitempty" gorm:"type:varchar(128);uniqueIndex,not null;comment:email"`

	// 安全信息 (JSON 忽略)
	Password string `json:"-" gorm:"type:varchar(128);not null;comment:encrypted psw"`
	// 1: 普通用户, 999: 超级管理员
	Role   int     `json:"role" gorm:"default:1;comment:1-User 999-Admin"`
	Status int     `json:"status" gorm:"type:smallint;default:1;comment:1-Active(正常) 2-Disabled(禁用)"` // Author作者档案
	Author *Author `gorm:"foreignKey:AdminID"`
}

// SetPassword 设置密码
// 功能：将明文密码加密后存入 User.Password 字段
func (m *Admin) SetPassword(password string) error {
	// GenerateFromPassword 使用 bcrypt 算法生成哈希
	// bcrypt.DefaultCost 默认值是 10，数值越高安全性越高，但性能消耗越大
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	m.Password = string(bytes)
	return nil
}

// CheckPassword 校验密码
// 功能：比对明文密码和数据库中存储的哈希值是否匹配
// 返回：true 表示密码正确，false 表示错误
func (m *Admin) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(m.Password), []byte(password))
	return err == nil
}
