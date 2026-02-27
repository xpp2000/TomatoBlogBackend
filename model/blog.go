package model

import "time"

type Post struct {
	BaseModel

	// base info
	Title    string `json:"title" gorm:"type:varchar(100);not null;comment:Title"`
	Subtitle string `json:"subtitle" gorm:"type:varchar(150);comment:Subtitle"`

	// SEO fields
	// Slug: for URL (e.g. /blog/how-to-design-db)
	Slug     string `json:"slug" gorm:"type:varchar(100);uniqueIndex;not null;comment:URL Alias"`
	Summary  string `json:"summary" gorm:"type:varchar(255);comment:Summary/SEO Description"`
	Keywords string `json:"keywords" gorm:"type:varchar(100);comment:SEO Keywords"`

	// 内容 (Markdown)
	// 使用 longtext 存储大量文字，MySQL 中 text 最多 64kb，longtext 可达 4GB
	Content string `json:"content" gorm:"type:text;not null;comment:Markdown Text"`

	// 视觉相关
	// 列表页展示的封面图，不是正文里的图w
	Cover string `json:"cover" gorm:"type:varchar(255);comment:Cover Image URL"`

	// 关联作者 (Belongs To)
	// 这里的 Admin 对应你之前的 Admin/User 结构体
	AuthorID uint64 `json:"author_id" gorm:"index;comment:Author ID"`
	Author   Author `json:"author" gorm:"foreignKey:AuthorID"`

	// 关联分类 (Belongs To)
	// 一篇文章通常属于一个主分类
	CategoryID uint64   `json:"category_id" gorm:"index;comment:Category ID"`
	Category   Category `json:"category" gorm:"foreignKey:CategoryID"`

	// 关联标签 (Many To Many)
	// 一篇文章可以有多个标签
	Tags []*Tag `json:"tags" gorm:"many2many:post_tags;"`

	// Status Control
	Status      int       `json:"status" gorm:"type:smallint;default:0;comment:0-draft 1-published 2-hided"`
	PublishedAt time.Time `json:"published_at" gorm:"comment:Published Date"`
	ReadCount   uint64    `json:"read_count" gorm:"default:0;comment:Read Count"`
}

// Category table
type Category struct {
	ID          uint64    `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null;unique;comment:Category Name"`
	Description string    `json:"description" gorm:"type:varchar(50);index;comment:Category Alias"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Tag table
type Tag struct {
	ID          uint64    `json:"id" gorm:"primarykey"`
	Name        string    `json:"name" gorm:"type:varchar(50);not null;unique;comment:Tag Name"`
	Description string    `json:"description" gorm:"type:varchar(50);index;comment:Table Alias"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Author struct {
	BaseModel
	Name     string `json:"name" gorm:"type:varchar(50);not null;unique;comment:Author Name"`
	Position string `json:"position" gorm:"type:varchar(64);not null;comment:Position"`
	Avatar   string `json:"avatar,omitempty" gorm:"type:varchar(255);comment:Avatar URL"`
	Mobile   string `json:"mobile,omitempty" gorm:"type:char(11);uniqueIndex;not null;comment:+86 phone number"`
	Email    string `json:"email,omitempty" gorm:"type:varchar(128);uniqueIndex;comment:email"`
}
