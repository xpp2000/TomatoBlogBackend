package api

import (
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"
	"tomatoBlogDB/service"

	"github.com/kataras/iris/v12"
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
// @Accept json
// @Produce json
// @Param data body dto.AdminLoginReq true "登录请求参数"
// @Router /api/v1/login [post]
func (m *AdminApi) Login(ctx iris.Context) {
	var req dto.AdminLoginReq

	// 1. set context
	m.SetContext(ctx)
	// 2. pass to BuildRequest
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 3. pass to Service
	token, adminUser, err := m.Service.Login(req)
	if err != nil {
		m.Fail(model.ResponseJson{
			Msg: err.Error(),
		})
		return
	}

	// 4. Login successfully
	m.Ok(model.ResponseJson{
		Msg: "Login successfully",
		Data: map[string]any{
			"token":     token,
			"user_info": adminUser,
		},
	})
}

// @Tags Admin
// @Summary Create normal user (writer)
// @Param Authorization header string true "Bearer Token"
// @Param body body dto.AdminAddReq true "Admin Info"
// @Router /api/v1/private/admin [post]
func (m *AdminApi) AddAdmin(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.AdminAddReq

	// 1. bind req
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 2. get current user Role
	role := ctx.Values().GetIntDefault("current_user_role", 0)

	// 3. call Service
	err := m.Service.CreateAdmin(req, role)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to create a new Admin " + err.Error()})
	}
	m.Ok(model.ResponseJson{Msg: "create a new Admin successfully"})
}

// @Tags Admin
// @Summary update admin status
// @Description only allow supreme admin(Role=999) do this. 1--active 2--disabled.
// @Param Authorization header string true "Bearer Token"
// @Param id path int true "admin or writer ID"
// @Param data body dto.AdminStatusReq true "status"
// @Router /api/v1/private/admin/{id}/status [patch]
func (m *AdminApi) UpdateStatus(ctx iris.Context) {
	// 1. permission check
	m.SetContext(ctx)
	role := ctx.Values().GetIntDefault("current_user_role", 1)
	if role != 999 {
		m.Fail(model.ResponseJson{Msg: "Permission denied: strictly for super admin"})
		return
	}

	// 2. bind req
	targetID, err := ctx.Params().GetUint64("id")
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "Invalid ID format " + err.Error()})
		return
	}
	var req dto.AdminStatusReq
	if err := ctx.ReadJSON(&req); err != nil {
		m.Fail(model.ResponseJson{Msg: "Invalid request body " + err.Error()})
		return
	}
	operatorID := ctx.Values().GetUint64Default("current_user_id", 0)
	// 3. call service
	if err := m.Service.UpdateAdminStatus(targetID, req.Status, operatorID); err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to update admin status " + err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "update Admin status successfully"})
}

// @Tag Admin
// Summary Delete a admin
// @Description only allow supreme admin(Role=999) do this.
// @Param id path int true "admin or writer ID"
// Router /api/v1/private/admin/{id} [delete]
func (m *AdminApi) DeleteAdmin(ctx iris.Context) {
	// 1. permission check
	m.SetContext(ctx)
	role := ctx.Values().GetIntDefault("role", 1)
	if role != 999 {
		m.Fail(model.ResponseJson{Msg: "Permission denied: strictly for super admin"})
		return
	}

	// 2. bind req
	targetID, err := ctx.Params().GetUint64("id")
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "Invalid ID format"})
		return
	}
	operatorID := ctx.Values().GetUint64Default("current_user_id", 0)

	// 3. call service
	if err := m.Service.DeleteAdmin(targetID, operatorID); err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to delete the admin"})
		return
	}
	m.Ok(model.ResponseJson{Msg: "delete Admin status successfully"})
}

func (m *AdminApi) ListAdmin(ctx iris.Context) {
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
