package dto

type Pagination struct {
	Page     int `json:"page" form:"page" validate:"required,gte=1"`
	PageSize int `json:"page_size" form:"page_size" validate:"required,gte=1,lte=100""`
}
