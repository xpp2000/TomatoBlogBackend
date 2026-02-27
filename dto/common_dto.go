package dto

type CommonIDReq struct {
	ID uint `param:"id" json:"id" form:"id" validate: "required,gt=0"`
}

// PageReq 通用分页请求参数
type PageReq struct {
	Page     int `json:"page" form:"page" validate:"omitempty,min=1"`
	PageSize int `json:"page_size" form:"page_size" validate:"omitempty,max=100"`
}

// GetPage return checked Page
func (r *PageReq) GetPage() int {
	if r.Page <= 0 {
		return 1
	}
	return r.Page
}

// GetPageSize return checked PageSize
func (r *PageReq) GetPageSize() int {
	if r.PageSize <= 0 {
		return 10
	}
	return r.PageSize
}
