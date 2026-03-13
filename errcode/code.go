package errcode

import "net/http"

// ==== conserve code ====
// 40001 'URL para invalid: '

var (
	// = others
	ErrPermissionDenied = &AppError{
		HttpCode: http.StatusForbidden,
		BizCode:  40300,
		Msg:      "Permission denied"}

	ErrURLParamInvalid = &AppError{
		HttpCode: http.StatusBadRequest,
		BizCode:  40001,
		Msg:      " Invalid URL parameter(s)"}

	/* --------------- Admin start --------------- */
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
	/* --------------- Admin end --------------- */

	/* --------------- Post start --------------- */
	ErrPostNotFound         = NewBizErr(4000201, "The post doesn't exist")
	ErrAuthorNotMatch       = NewBizErr(4000202, "Permission denied, current author doesn't match")
	ErrPostLocked           = NewBizErr(4000203, "the post has been locked, please let admin unlock it")
	ErrCategoryNotFound     = NewBizErr(4000204, "Undefined category")
	ErrCategoryDeleteRefuse = NewBizErr(4000205, "fail to delete: there are posts associated with this category")

	/* --------------- Post end --------------- */

	/* --------------- Author start --------------- */
	ErrAuthorNotFound = NewBizErr(4000301, "The Author doesn't exists")

	/* --------------- Author end --------------- */

)
