package errcode

import "fmt"

type AppError struct {
	HttpCode int    // 对应的 HTTP 状态码 (如 400, 403, 404, 500)
	BizCode  int    // 业务逻辑码 (如 43011)
	Msg      string // 返回给前端的提示语
	RawError error  // 真实底层报错，通常系统报错才记录
}

func (e *AppError) Error() string {
	if e.RawError != nil {
		return fmt.Sprintf("[BizCode: %d] %s (Raw: %v)", e.BizCode, e.Msg, e.RawError)
	}
	return fmt.Sprintf("[BizCode: %d] %s", e.BizCode, e.Msg)
}

// 实现 Unwrap 接口，完美兼容 Go 原生的 errors.Is 和 errors.As
func (e *AppError) Unwrap() error {
	return e.RawError
}

// NewBizErr 专门用于抛出业务逻辑错误 (默认 HTTP 400)
func NewBizErr(bizCode int, msg string) *AppError {
	return &AppError{
		HttpCode: 400,
		BizCode:  bizCode,
		Msg:      msg,
	}
}

// NewSysErr 专门用于抛出系统崩溃错误 (默认 HTTP 500)
func NewSysErr(rawErr error) *AppError {
	return &AppError{
		HttpCode: 500,
		BizCode:  50000,
		Msg:      "服务器内部异常，请稍后再试",
		RawError: rawErr,
	}
}
