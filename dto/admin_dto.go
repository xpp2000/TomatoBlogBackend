package dto

// ==== AdminLoginReq
type AdminLoginReq struct {
	Name     string `json:"name" form:"name" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}

type AdminAddReq struct {
	Name     string `json:"name" form:"name" validate:"required,min=3,max=20"`
	RealName string `json:"real_name" form:"real_name" validate:"omitempty,max=20"`
	Mobile   string `json:"mobile" form:"mobile" validate:"omitempty,len=11"` // 中国手机号通常11位
	Email    string `json:"email" form:"email" validate:"omitempty,email"`    // 校验邮箱格式
	Password string `json:"password" form:"password" validate:"required,min=6"`

	// Avatar 这个字段通常不由前端直接传 JSON 过来，而是文件上传后由 Controller 赋值路径
	// 所以标记为 omitempty，且 json 忽略
	Avatar string `json:"avatar,omitempty" form:"avatar"`
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
