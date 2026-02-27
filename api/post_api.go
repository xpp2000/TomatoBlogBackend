package api

import (
	"tomatoBlogDB/dto"
	"tomatoBlogDB/model"
	"tomatoBlogDB/service"

	"github.com/kataras/iris/v12"
)

type PostApi struct {
	*BaseApi
	Service *service.PostService
}

func NewPostApi() *PostApi {
	return &PostApi{
		BaseApi: NewBaseApi(),
		Service: service.NewPostService(),
	}
}

type CategoryApi struct {
	*BaseApi
	Service *service.CategoryService
}

func NewCategoryApi() *CategoryApi {
	return &CategoryApi{
		BaseApi: NewBaseApi(),
		Service: service.NewCategoryService(),
	}
}

/* ===== POST API begin ===== */

// @Summary Get post details
// @Param slug_or_id path string true "Post ID or Slug"
// @Router /api/v1/post/{slug_or_id} [get]
func (m *PostApi) GetPost(ctx iris.Context) {
	m.SetContext(ctx)
	param := ctx.Params().Get("slug_or_id")

	post, err := m.Service.GetPostDetail(param)
	if err != nil {
		m.Fail(model.ResponseJson{Code: 404, Msg: "post dose't exist"})
		return
	}
	m.Ok(model.ResponseJson{Data: post})
}

// @Tags Post
// @Summary Create a new post
// @Param Authorization header string true "Bearer Token"
// @Param body body dto.PostAddReq true "Post Info"
// @Router /api/v1/private/post [post]
func (m *PostApi) AddPost(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.PostAddReq

	// 1. bind req
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 2. get current user ID
	authorID := ctx.Values().GetUint64Default("current_user_id", 0)
	if authorID == 0 {
		// - as expected, middleware has already blocked
		m.Fail(model.ResponseJson{Code: 401, Msg: "Cannot acquire User Identity"})
		return
	}

	// 3. call Service(pass into sate authorID)
	err := m.Service.CreatePost(req, authorID)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to publish " + err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "publish successfully"})
}

// @Tags Post
// @Summary Update a post
// @Param Authorization header string true "Bearer Token"
// @Param id path int true "Post ID"
// @Param body body dto.PostUpdateReq true "Update Info"
// @Router /api/v1/private/post/{id} [put]
func (m *PostApi) UpdatePost(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.PostUpdateReq
	// 1. get Post ID
	id, err := ctx.Params().GetInt64("id")
	if err != nil || id <= 0 {
		m.Fail(model.ResponseJson{Code: 400, Msg: "invalid post ID"})
		return
	}

	// 2. bind parameters in body
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}
	req.ID = uint64(id)

	// 3. make sure that JWT has been phrased
	operatorID := ctx.Values().GetUint64Default("current_user_id", 0)
	operatorRole := ctx.Values().GetIntDefault("current_user_role", 0)
	if operatorID == 0 {
		m.Fail(model.ResponseJson{Code: 401, Msg: "please login first"})
		return
	}

	// 4. call service
	err = m.Service.UpdatePost(req, operatorID, operatorRole)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: "fail to update post: " + err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "update successfully"})
}

// @Summary Change post status
// @Param Authorization header string true "Bearer Token"
// @Param id path int true "Post ID"
// @Param body body dto.PostStatusReq true "Status"
// @Router /api/v1/private/post/{id}/status [patch]
func (m *PostApi) UpdateStatus(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.PostStatusReq

	// 1. fetch ID
	id, _ := ctx.Params().GetInt64("id")

	// 2. bind Body
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}
	req.ID = uint64(id)

	// 3. fetch role
	uid := ctx.Values().GetUint64Default("current_user_id", 0)
	role := ctx.Values().GetIntDefault("current_user_role", 0)

	// 4. call service
	err := m.Service.UpdateStatus(req, uid, role)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "change status successfully"})
}

// @Summary Get post list
// @Param page query int false "Page Number"
// @Param page_size query int false "Page Size"
// @Param keyword query string false "Search Keyword"
// @Router /api/v1/posts [get]
func (m *PostApi) ListPosts(ctx iris.Context) {
	m.SetContext(ctx)

	var req dto.PostListReq
	// 1. BindQuery
	if err := ctx.ReadQuery(&req); err != nil {
		return
	}

	// default page parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 2. service
	posts, total, err := m.Service.GetPostList(req)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: err.Error()})
		return
	}

	// 3. successfully
	m.Ok(model.ResponseJson{
		Data: iris.Map{
			"list":  posts,
			"total": total,
			"page":  req.Page,
		},
	})
}

// @Summary Delete a post
// @Param Authorization header string true "Bearer Token"
// @Param id path int true "Post ID"
// @Router /api/v1/private/post/{id} [delete]
func (m *PostApi) DeletePost(ctx iris.Context) {
	m.SetContext(ctx)

	id, _ := ctx.Params().GetInt64("id")

	uid := ctx.Values().GetUint64Default("current_user_id", 0)
	role := ctx.Values().GetIntDefault("current_user_role", 0)

	// Service 内部逻辑：
	// 1. GetPostByID
	// 2. if role != Admin && post.AuthorID != uid { return error }
	// 3. db.Delete(&post) // GORM 默认是软删除
	err := m.Service.DeletePost(uint64(id), uid, role)

	if err != nil {
		m.Fail(model.ResponseJson{Msg: err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "删除成功"})
}

/* ===== POST API end ===== */

// @Summary Get category details
// @Param name_or_id path string true "Post ID or Name"
// @Router /api/v1/category/{name_or_id} [get]
func (m *CategoryApi) GetCategory(ctx iris.Context) {
	m.SetContext(ctx)
	param := ctx.Params().Get("name_or_id")

	post, err := m.Service.GetCategoryDetail(param)
	if err != nil {
		m.Fail(model.ResponseJson{Code: 404, Msg: "post dose't exist"})
		return
	}
	m.Ok(model.ResponseJson{Data: post})
}

// @Summary Get category list
// @Param page query int false "Page Number"
// @Param page_size query int false "Page Size"
// @Router /api/v1/categories [get]
func (m *CategoryApi) ListCategory(ctx iris.Context) {
	m.SetContext(ctx)

	var req dto.CategoryListReq
	// 1. BindQuery
	if err := ctx.ReadQuery(&req); err != nil {
		return
	}

	// default page parameters
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 2. service
	cates, total, err := m.Service.GetCategoryList(req)
	if err != nil {
		m.Fail(model.ResponseJson{Msg: err.Error()})
		return
	}

	// 3. successfully
	m.Ok(model.ResponseJson{
		Data: iris.Map{
			"list":  cates,
			"total": total,
			"page":  req.Page,
		},
	})
}
