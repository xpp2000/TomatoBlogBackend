package errcode

import "net/http"

var (
	// = 400级别
	ErrEmailExist         = NewBizErr(4000101, "Email has existed")
	ErrMobileExist        = NewBizErr(4000102, "Phone number has been registered")
	ErrNameExist          = NewBizErr(4000103, "Name has existed")
	ErrRoleTargetDeny     = NewBizErr(4000104, "Can create that target role")
	ErrPasswordWrong      = NewBizErr(4000105, "Fail to set password")
	ErrMaliciousTarget    = NewBizErr(4000108, "Wrong Target ID")
	ErrMaliciousOperation = NewBizErr(4000109, "Current operator can't do this")
	ErrAdminNotFound      = NewBizErr(4000106, "The admin doesn't exist")
	ErrPassWordIncorrect  = NewBizErr(4000107, "The password is incorrect, pleas check again")
	ErrAdminStatus        = NewBizErr(4000110, "Status can only be 1 | 2")
	ErrDeletePostFirst    = NewBizErr(4000119, "Fail to delete admin or writer, please delete his/her posts first")
	// = others
	ErrPermissionDenied = &AppError{
		HttpCode: http.StatusForbidden,
		BizCode:  40300,
		Msg:      "Permission denied"}
)
