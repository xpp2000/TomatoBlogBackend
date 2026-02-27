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

func NewAdminApi() *AdminApi {
	return &AdminApi{
		BaseApi: NewBaseApi(),
		Service: service.NewAdminService(),
	}
}

// @Tags Post
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
