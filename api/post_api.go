package api

import (
	"tomatoBlogDB/dto"
	"tomatoBlogDB/errcode"
	"tomatoBlogDB/model"
	"tomatoBlogDB/service"

	"github.com/kataras/iris/v12"
)

type PostApi struct {
	*BaseApi
	Service *service.PostService
}

// NewPostApi 强制外部注入装配好的 Service
func NewPostApi(svc *service.PostService) *PostApi {
	return &PostApi{
		BaseApi: NewBaseApi(),
		Service: svc,
	}
}

type CategoryApi struct {
	*BaseApi
	Service *service.CategoryService
}

func NewCategoryApi(svc *service.CategoryService) *CategoryApi {
	return &CategoryApi{
		BaseApi: NewBaseApi(),
		Service: svc,
	}
}

type AuthorApi struct {
	*BaseApi
	Service *service.AuthorService
}

func NewAuthorApi(svc *service.AuthorService) *AuthorApi {
	return &AuthorApi{
		BaseApi: NewBaseApi(),
		Service: svc,
	}
}

/* ===== POST API begin ===== */
// @Tags Post
// @Summary Get post details
// @Description ** url参数可以为slug或id **
// @Param slug_or_id path string true "Post ID or Slug"
// @Produce json
// @Success 200 200 {object} model.ResponseJson{Data=dto.PostSimple} "成功"
// @Router /api/v1/post/{slug_or_id} [get]
func (m *PostApi) GetPost(ctx iris.Context) {
	m.SetContext(ctx)
	param := ctx.Params().Get("slug_or_id")

	post, err := m.Service.GetPostDetail(param)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Data: post})
}

// @Tags Post
// @Summary Create a new post
// @Description - 可以指定slug，但建议置为空，让程序自动生成
// @Description - published_at: 发布时间,RFC3339格式："2006-01-02T15:04:05Z07:00" 。若留空则默认为当前时间。
// @Description - 后门字段target_author_id,仅当超级管理员时生效
// @Accept json
// @Param Authorization header string false "Bearer Token"
// @Param body body dto.PostAddReq true "Post Info"
// @Success 200 200 {object} model.ResponseJson{Msg: "publish successfully" "成功"
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
	operatorRole := ctx.Values().GetIntDefault("current_user_role", 0)
	if authorID == 0 {
		// - as expected, middleware has already blocked
		m.HandleError(ctx, errcode.ErrAuthorNotMatch)
		return
	}

	// 3. call Service(pass into sate authorID)
	err := m.Service.CreatePost(req, authorID, operatorRole)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "publish successfully"})
}

// @Tags Post
// @Summary Update a post
// @Description 更新文章内容。支持部分更新（动态 PATCH 逻辑）：前端只需传入需要修改的字段，未传入的字段将保持原样。
// @Description 注意：不支持修改文章作者。仅限文章原作者或系统管理员操作。
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path int true "Post ID"
// @Param body body dto.PostUpdateReq true "Update Info(仅传需要修改的字段)"
// @Success 200 {object} model.ResponseJson "update successfully"
// @Router /api/v1/private/post/{id} [put]
func (m *PostApi) UpdatePost(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.PostUpdateReq
	// 1. get Post ID
	id, err := ctx.Params().GetInt64("id")
	if err != nil || id <= 0 {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
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
		m.HandleError(ctx, errcode.ErrPermissionDenied)
		return
	}

	// 4. call service
	err = m.Service.UpdatePost(req, operatorID, operatorRole)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "update successfully"})
}

// @Tags Post
// @Summary Change post status
// @Description Update status field
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path int true "Post ID"
// @Param body body dto.PostStatusReq true "Status"
// @Success 200 {object} model.ResponseJson "update successfully"
// @Router /api/v1/private/post/{id}/status [patch]
func (m *PostApi) UpdateStatus(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.PostStatusReq

	// 1. fetch ID
	id, errP := ctx.Params().GetInt64("id")
	if errP != nil || id <= 0 {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
		return
	}

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
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "change status successfully"})
}

// @Tags Post
// @Summary Get post list
// @Description 获取文章列表（支持分页）。
// @Description 注意：为了优化传输性能，列表接口返回的 Post 对象不包含 Content (文章正文)。
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认: 1)" default(1)
// @Param page_size query int false "每页数量 (默认: 10, 最大: 100)" default(10)
// @Param keyword query string false "Search Keyword"
// @Success 200 {object} model.ResponseJson{data=dto.PostListResp} "获取列表成功"
// @Router /api/v1/posts [get]
func (m *PostApi) ListPosts(ctx iris.Context) {
	m.SetContext(ctx)

	var req dto.PostListReq
	// 1. BindQuery
	if err := ctx.ReadQuery(&req); err != nil {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
		return
	}

	// default page parameters, can put these into service
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 2. service
	posts, total, err := m.Service.GetPostList(req)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}

	// 3. successfully
	m.Ok(model.ResponseJson{
		Data: dto.PostListResp{
			List:  posts,
			Total: total,
			Page:  req.Page,
		},
	})
}

// @Tags Post
// @Summary Delete a post
// @Description 删除指定的文章（注意：此操作为软删除，数据仍在数据库中，只是被标记为已删除）。
// @Description 权限要求：仅限文章原作者或系统管理员(Role=Admin)可以执行此操作。
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path int true "Post ID"
// @Success 200 {object} model.ResponseJson "delete Post status successfully / 删除成功"
// @Router /api/v1/private/post/{id} [delete]
func (m *PostApi) DeletePost(ctx iris.Context) {
	m.SetContext(ctx)

	id, errP := ctx.Params().GetInt64("id")
	if errP != nil {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
		return
	}

	uid := ctx.Values().GetUint64Default("current_user_id", 0)
	role := ctx.Values().GetIntDefault("current_user_role", 0)

	// Service 内部逻辑：
	// 1. GetPostByID
	// 2. if role != Admin && post.AuthorID != uid { return error }
	// 3. db.Delete(&post) // GORM 默认是软删除
	err := m.Service.DeletePost(uint64(id), uid, role)

	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "delete Post status successfully删除成功"})
}

/* ===== POST API end ===== */

/* ===== Category API begin ===== */

// @Tags Category
// @Summary Create a new category
// @Description 创建一个新分类。
// @Description 权限要求：仅限系统管理员 (Role=Admin) 可以执行此操作。
// @Accept json
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param body body dto.CategoryAddReq true "Post Info"
// @Success 200 {object} model.ResponseJson "create a new category successfully"
// @Router /api/v1/private/category [post]
func (m *CategoryApi) AddCategory(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.CategoryAddReq

	// 1. bind req
	if !m.BuildRequest(BuildRequestOption{DTO: &req, BindBody: true}) {
		return
	}

	// 2. get current user ID
	role := ctx.Values().GetIntDefault("current_user_role", 0)

	// 3. call Service(pass into sate authorID)
	err := m.Service.CreateCategory(req, role)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "create a new category successfully"})
}

// @Tags Category
// @Summary Get category details
// @Description 根据分类 ID 或分类名称获取分类详情。支持灵活查询
// @Produce json
// @Param name_or_id path string true "Category ID or Name (分类的 ID 或名称)"
// @Success 200 {object} model.ResponseJson "获取分类详情成功"
// @Router /api/v1/category/{name_or_id} [get]
func (m *CategoryApi) GetCategory(ctx iris.Context) {
	m.SetContext(ctx)
	param := ctx.Params().Get("name_or_id")

	post, err := m.Service.GetCategoryDetail(param)
	if err != nil {
		m.HandleError(ctx, errcode.ErrPostNotFound)
		return
	}
	m.Ok(model.ResponseJson{Data: post})
}

// @Tags Category
// @Summary Get category list
// @Description 分页获取分类列表
// @Produce json
// @Param page query int false "Page Number (页码，默认 1)" default(1)
// @Param page_size query int false "Page Size (每页数量，默认 10)" default(10)
// @Success 200 {object} model.ResponseJson{data=dto.CategoryListResp} "获取分类列表成功"
// @Router /api/v1/categories [get]
func (m *CategoryApi) ListCategory(ctx iris.Context) {
	m.SetContext(ctx)

	var req dto.CategoryListReq
	// 1. BindQuery
	if err := ctx.ReadQuery(&req); err != nil {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
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
		m.HandleError(ctx, err)
		return
	}

	// 3. successfully
	m.Ok(model.ResponseJson{
		Data: dto.CategoryListResp{
			List:  cates,
			Total: total,
			Page:  req.Page,
		},
	})
}

// @Tags Category
// @Summary Delete a category
// @Description 删除指定的分类
// @Description 权限要求：通常仅限系统管理员 (Role=Admin) 可以执行此操作。
// @Produce json
// @Param Authorization header string false "Bearer Token"
// @Param id path int true "Category ID"
// @Success 200 {object} model.ResponseJson "删除成功"
// @Router /api/v1/private/category/{id} [delete]
func (m *CategoryApi) DeleteCategory(ctx iris.Context) {
	m.SetContext(ctx)

	id, errP := ctx.Params().GetInt64("id")
	if errP != nil || id <= 0 {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
		return
	}
	role := ctx.Values().GetIntDefault("current_user_role", 0)
	err := m.Service.DeleteCategory(uint64(id), role)

	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Msg: "删除成功"})
}

/* ===== Category API end ===== */

/* ===== Author API begin ===== */

// @Tags Author
// @Summary Get author details
// @Description 根据作者 ID 或名称 (PenName/Username) 获取作者详情。支持灵活查询。
// @Produce json
// @Param name_or_id path string true "Author Name or ID"
// @Success 200 {object} model.ResponseJson "获取作者详情成功"
// @Router /api/v1/author/{name_or_id} [get]
func (m *AuthorApi) GetAuthor(ctx iris.Context) {
	m.SetContext(ctx)
	param := ctx.Params().Get("name_or_id")

	author, err := m.Service.GetAuthorDetail(param)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}
	m.Ok(model.ResponseJson{Data: author})
}

// @Tags Author
// @Summary Get author list
// @Description 分页获取作者列表
// @Produce json
// @Param page query int false "Page Number (页码，默认 1)" default(1)
// @Param page_size query int false "Page Size (每页数量，默认 10)" default(10)
// @Success 200 {object} model.ResponseJson{data=dto.AuthorListResp} "获取作者列表成功"
// @Router /api/v1/authors [get]
func (m *AuthorApi) ListAuthors(ctx iris.Context) {
	m.SetContext(ctx)
	var req dto.AuthorListReq

	// 1. BindQuery
	if err := ctx.ReadQuery(&req); err != nil {
		m.HandleError(ctx, errcode.ErrURLParamInvalid)
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
	authors, total, err := m.Service.GetAuthorList(req)
	if err != nil {
		m.HandleError(ctx, err)
		return
	}

	// 3. successfully
	m.Ok(model.ResponseJson{
		Msg: "successfully fetch the author list", // 顺手帮你加个友好的提示信息
		Data: dto.AuthorListResp{
			List:  authors,
			Total: total,
			Page:  req.Page,
		},
	})
}

/*
@Tags Author
@Summary Delete an author
@Param Authorization header string true "Bearer Token"
@Param id path int true "Author ID"
@Router /api/v1/private/author/{id} [delete]
func (m *AuthorApi) DeleteAuthor(ctx iris.Context) {
	m.SetContext(ctx)

	id, _ := ctx.Params().GetInt64("id")

	role := ctx.Values().GetIntDefault("current_user_role", 0)

	err := m.Service.DeleteAuthor(uint64(id), role)

	if err != nil {
		m.Fail(model.ResponseJson{Msg: err.Error()})
		return
	}
	m.Ok(model.ResponseJson{Msg: "删除成功"})
}
*/
/* ===== Author API end ===== */
