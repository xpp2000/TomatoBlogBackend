package dto

/* Post begin */
type PostAddReq struct {
	Title    string `json:"title" validate:"required,max=100"`
	Subtitle string `json:"subtitle" validate:"omitempty,max=150"`
	Slug     string `json:"slug" validate:"omitempty,max=100"`
	Summary  string `json:"summary" validate:"omitempty,max=255"`
	Keywords string `json:"keywords" validate:"omitempty,max=255"`

	Content string `json:"content" validate:"required"`
	Cover   string `json:"cover" validate:"omitempty,url"`

	CategoryID  uint64   `json:"category_id" validate:"required,gt=0"`
	Tags        []string `json:"tags" validate:"omitempty"`         //
	PublishedAt string   `json:"published_at" validate:"omitempty"` // RFC3339 "published_at": "2026-02-27T18:00:00+08:00"
}

type PostUpdateReq struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title" validate:"required,max=100"`
	Subtitle string `json:"subtitle" validate:"omitempty,max=150"`
	Slug     string `json:"slug" validate:"omitempty,max=100"`
	Summary  string `json:"summary" validate:"omitempty,max=255"`
	Keywords string `json:"keywords" validate:"omitempty,max=255"`
	Content  string `json:"content" validate:"required"`
	Cover    string `json:"cover" validate:"omitempty,url"`
	// 分类通常是必填的，即使不改也要传原值
	CategoryID uint64 `json:"category_id" validate:"required,gt=0"`
	// 标签 ID 列表 (全量覆盖)
	Tags []string `json:"tags" validate:"omitempty"`
	// 状态 (例如想把已发布改为草稿)
	Status int `json:"status" validate:"oneof=0 1 2"`
}

type PostStatusReq struct {
	ID     uint64 `json:"id" validate:"required,gt=0"`
	Status int    `json:"status" validate:"required,oneof=0 1 2"`
}

type PostListReq struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"page_size" form:"page_size"`
	Keyword    string `json:"keyword" form:"keyword"`
	CategoryID uint64 `json:"category_id" form:"category_id"`
	TagID      uint64 `json:"tag_id" form:"tag_id"`
	Status     *int   `json:"status" form:"status"` // point type, allow 0/nil values
}

/* ===== Post end ===== */

/* ===== Category begin ===== */
type CategoryListReq struct {
	Pagination
}

type CategoryAddReq struct {
	Name        string `json:"name" validate:"required,max=128"`
	Description string `json:"description" validate:"omitempty,max=128"`
}

/* ===== Category end ===== */

/* ===== Author begin ===== */
type AuthorAddReq struct {
	Name     string `json:"name" validate:"required,max=128"`
	Position string `json:"position" validate:"required,max=128"`
	Avatar   string `json:"avatar,omitempty" validate:"required,url"`
	Mobile   string `json:"mobile,omitempty" validate:"omitempty,max=32"`
	Email    string `json:"email,omitempty" validate:"omitempty,max=128"`
}

type AuthorListReq struct {
	Pagination
}

/* ===== Author end ===== */

/* ===== View dto begin ====== */
type CategorySimple struct {
	ID          uint64 `json:"id" `
	Name        string `json:"name" `
	Description string `json:"description"`
}

type AuthorSimple struct {
	ID       uint64 `json:"id" `
	PenName  string `json:"pen_name" `
	Position string `json:"position"`
	Avatar   string `json:"avatar,omitempty"`
	Mobile   string `json:"mobile,omitempty"`
	Email    string `json:"email,omitempty" `
}

type PostSimple struct {
	ID       uint64 `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`

	Slug     string `json:"slug" `
	Summary  string `json:"summary" `
	Keywords string `json:"keywords" `

	Content string `json:"content"`

	// 视觉相关
	// 列表页展示的封面图，不是正文里的图w
	Cover string `json:"cover"`

	// 关联作者 (Belongs To)
	// 这里的 Admin 对应你之前的 Admin/User 结构体
	AuthorID uint64       `json:"author_id"`
	Author   AuthorSimple `json:"author"`

	// 关联分类 (Belongs To)
	// 一篇文章通常属于一个主分类
	CategoryID   uint64 `json:"category_id"`
	CategoryName string `json:"category_name"`

	// 关联标签 (Many To Many)
	// 一篇文章可以有多个标签
	Tags []string `json:"tags" `

	// Status Control
	Status      int    `json:"status" `
	PublishedAt string `json:"published_at" `
	ReadCount   uint64 `json:"read_count" `
}

/* ===== View dto end ====== */
