package api

import (
	"time"
	"tomatoBlogDB/dto"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/model"
	"tomatoBlogDB/service"

	"github.com/kataras/iris/v12"
)

// 最好单独成包pkg/errcode/code.go
const (
	ERROR_CODE_ADD            = 401001
	ERROR_CODE_DELETE         = 401002
	ERROR_CODE_WRONGLOGININFO = 401003
	ERROR_CODE_ACCOUNTBANED   = 401004
	ERROR_CODE_SUPREMEONLY    = 401005
)

type AdminApi struct {
	*BaseApi
	Service *service.AdminService
}

func NewAdminApi(adminService *service.AdminService) *AdminApi {
	return &AdminApi{
		BaseApi: NewBaseApi(),
		Service: adminService,
	}
}

// @Tags Admin
// @Summary Login
// @Description
// @Description **业务错误码 (Code) 说明：**
// @Description - `401003` - 用户名或密码错误
// @Description - `401004` - 该账号已被冻结，禁止登录
// @Accept json
// @Produce json
// @Param data body dto.AdminLoginReq true "登录请求参数"
// @Success 200 {object} model.ResponseJson{data=dto.AdminLoginResp} "成功"
// @Router /api/v1/login [post]
func (m *AdminApi) Login(ctx iris.Context) {
	// 1. set context
	m.SetContext(ctx)
	var req dto.AdminLoginReq
	// 2. pass to BuildRequest
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 3. pass to Service
	token, adminUser, err := m.Service.Login(req)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	var role string
	switch adminUser.Role {
	case 999:
		role = "admin"
	case 1:
		role = "user"
	}
	// 4. Login successfully, Dual token
	ctx.SetCookieKV("access_token", token, iris.CookieExpires(time.Hour*24),
		iris.CookieHTTPOnly(true),
	)
	m.Ok(model.ResponseJson{
		Msg: "Login successfully",
		Data: map[string]any{
			"token": token,
			"role":  role,
		},
	})
}

// @Tags Admin
// @Summary Create normal user (writer)
// @Description ** 任意添加admin表【！！非业务】**
// @Description ** ---- 业务错误码（Cpde）说明 ---- **
// @Description -401005
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param body body dto.AdminAddReq true "Admin Info"
// @Success 200 {object} model.ResponseJson "成功响应。注意：实际返回值为 Code: 20100, Msg: 'create a new Admin successfully'"
// @Router /api/v1/private/admin [post]
func (m *AdminApi) AddAdmin(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.AdminAddReq

	// 0. authorization check
	role := ctx.Values().GetIntDefault("current_user_role", 0)
	if role != 999 { // 假设 999 是超级管理员
		// 在这里拦截，清清楚楚地返回 403 Forbidden
		err1 := errcode.ErrPermissionDenied
		m.HandleError(ctx, err1)
		return
	}

	// 1. bind req
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 2. get current user Role

	// 3. call Service
	err := m.Service.CreateAdmin(req, role)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Code: 20100,
		Msg: "create a new Admin successfully"})
}

// @Tags Admin
// @Summary update an admin status
// @Description only allow supreme admin(Role=999) do this. 1--active 2--disabled.
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path string true "admin ID"
// @Param data body dto.AdminStatusReq true "status"
// @Success 200 200 {object} model.ResponseJson "成功响应。实际返回值为 Msg: 'update Admin status successfully'}"
// @Router /api/v1/private/admin/{id}/status [patch]
func (m *AdminApi) UpdateStatus(ctx iris.Context) {
	// 1. permission check
	m.SetContext(ctx)
	role := ctx.Values().GetIntDefault("current_user_role", 1)
	if role != 999 {
		m.HandleError(ctx, errcode.ErrPermissionDenied)
		return
	}

	// 2. bind req
	targetID, err := ctx.Params().GetUint64("id")
	if err != nil {
		m.HandleError(ctx, errcode.NewBizErr(40001, "URL para invalid: "+err.Error()))
		return
	}
	var req dto.AdminStatusReq
	if err := ctx.ReadJSON(&req); err != nil {
		m.HandleError(ctx, err)
		return
	}
	operatorID := ctx.Values().GetUint64Default("current_user_id", 0)
	// 3. call service
	if err := m.Service.UpdateAdminStatus(targetID, req.Status, operatorID); err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "update Admin status successfully"})
}

// @Tags Admin
// @Summary Delete an admin
// @Description only allow supreme admin(Role=999) do this.
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path int true "admin ID"
// @Success 200 200 {object} model.ResponseJson "成功响应。实际返回值为 Msg: 'delete Admin status successfully'"
// @Router /api/v1/private/admin/{id} [delete]
func (m *AdminApi) DeleteAdmin(ctx iris.Context) {
	// 1. permission check
	m.SetContext(ctx)
	role := ctx.Values().GetIntDefault("current_user_role", 1)
	if role != 999 {
		m.HandleError(ctx, errcode.ErrPermissionDenied)
		return
	}

	// 2. bind req
	targetID, err := ctx.Params().GetUint64("id")
	if err != nil {
		m.HandleError(ctx, errcode.NewBizErr(40001, "URL para invalid: "+err.Error()))
		return
	}
	operatorID := ctx.Values().GetUint64Default("current_user_id", 0)

	// 3. call service
	if err := m.Service.DeleteAdmin(targetID, operatorID); err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "delete Admin status successfully"})
}

// @Tags Admin
// @Summary Get admin list
// @Param Authorization header string true "Bearer Token"
// @Param page query int false "Page Number"
// @Param page_size query int false "Page Size"
// @Router /api/v1/private/admins [get]
func (m *AdminApi) ListAdmins(ctx iris.Context) {
	m.SetContext(ctx)
	// 1. permission check
	role := ctx.Values().GetIntDefault("current_user_role", 1)
	if role != 999 {
		m.Fail(model.ResponseJson{Msg: "Permission denied: strictly for super admin"})
		return
	}

	// 2. bind req (GET parameters)
	var req dto.AdminListReq
	if err := ctx.ReadQuery(&req); err != nil {
		m.Fail(model.ResponseJson{Msg: "Invalid query parameters: " + err.Error()})
		return
	}
	// 3. call service
	list, total, err := m.Service.GetAdminList(req)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to get admin list: " + err.Error()})
		return
	}

	// 4. return JSON
	m.Ok(model.ResponseJson{
		Msg: "get admin list successfully",
		Data: iris.Map{
			"list":  list,
			"total": total,
		},
	})
}
