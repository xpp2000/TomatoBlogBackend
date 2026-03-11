package dto

// ==== AdminLoginReq
type AdminLoginReq struct {
	Name     string `json:"name" form:"name" validate:"required"`
	Password string `json:"password" form:"password" validate:"required" binding:"required"`
}
type AdminLoginResp struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1Ni... "` // 可以顺手加个 example
	Role  string `json:"role" example:"admin"`
}

type AdminAddReq struct {
	Name     string `json:"name" form:"name" validate:"required,min=3,max=20"`
	RealName string `json:"real_name" form:"real_name" validate:"omitempty,max=20"`
	Mobile   string `json:"mobile" form:"mobile" validate:"required,len=11"` // 中国手机号通常11位
	Email    string `json:"email" form:"email" validate:"required,email"`    // 校验邮箱格式
	Password string `json:"password" form:"password" validate:"required,min=6"`

	Role    int    `json:"role" form:"role" validate:"required,max=999"`
	PenName string `json:"pen_name" form:"pen_name" validate:"required,max=64"`
}

// ==== Update
type AdminUpdateReq struct {
	ID       uint   `json:"id" form:"id"`
	ReadName string `json:"real_name" form:"real_name" validate:"omitempty,max=20"`
	Mobile   string `json:"mobile" form:"mobile" validate:"omitempty,len=11"`
	Email    string `json:"email" form:"email" validate:"omitempty,empty"`
}

// ==== List
type UserListReq struct {
	PageReq        // 嵌入通用分页
	Name    string `json:"name" form:"name"` // 支持按用户名模糊搜索
}

// ==== Status Update
type AdminStatusReq struct {
	Status int `json:"status" validate:"oneof=1 2"`
}

// ====
type AdminListReq struct {
	PageReq
	Keyword string `url:"keyword" form:"keyword"` // 搜索关键字 (用户名或手机号)
	Role    *int   `url:"role" form:"role"`       // 筛选角色
	Status  *int   `url:"status" form:"status"`   // 筛选状态
}

/* ===== Response begin ===== */
// AdminListResp 扁平化的返回结构 (专为前端 TanStack Table 优化)
type AdminListResp struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Mobile    string `json:"mobile"`
	Email     string `json:"email"`
	Role      int    `json:"role"`
	Status    int    `json:"status"`
	CreatedAt string `json:"created_at"`

	// 提前把 Author 表里的展示信息提出来，绝不返回 Password 等敏感信息
	PenName   string `json:"pen_name"`
	Avatar    string `json:"avatar"`
	PenEmail  string `json:"pen_email"`
	PenMobile string `json:"pen_mobile"`
}
