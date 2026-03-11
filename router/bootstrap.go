/* used for DI */
package router

import (
	"tomatoBlogDB/api"
	"tomatoBlogDB/dao"
	"tomatoBlogDB/service"

	"gorm.io/gorm"
)

// AppContainer 依赖注入的大容器，里面装满了全站唯一的实例
type AppContainer struct {
	PostApi     *api.PostApi
	CategoryApi *api.CategoryApi
	AdminApi    *api.AdminApi
	AuthorApi   *api.AuthorApi
	// ... 可以继续加 AdminApi 等
}

// InitDI 初始化整个依赖注入树
// 关键点：要求外部必须把已经连好的 *gorm.DB 传进来！
func InitDI(db *gorm.DB) *AppContainer {

	/* ===== 1. DB & Base ===== */
	baseDao := dao.NewBaseDao(db)

	/* ===== 2. DAO ===== */
	postDao := dao.NewPostDao(baseDao)
	tagDao := dao.NewTagDao(baseDao)
	categoryDao := dao.NewCategoryDao(baseDao)
	adminDao := dao.NewAdminDao(baseDao)
	authorDao := dao.NewAuthorDao(baseDao)
	// adminDao := dao.NewAdminDao(baseDao)

	/* ===== 3. SERVICE ===== */
	postService := service.NewPostService(postDao, tagDao) // 假设你需要这些
	categoryService := service.NewCategoryService(categoryDao, postDao)
	adminService := service.NewAdminService(adminDao, postDao)
	authorService := service.NewAuthorService(authorDao, postDao)
	/* ===== 4. API ===== */
	postApi := api.NewPostApi(postService)
	categoryApi := api.NewCategoryApi(categoryService)
	adminApi := api.NewAdminApi(adminService)
	authorApi := api.NewAuthorApi(authorService)
	// 返回装配好的容器
	return &AppContainer{
		PostApi:     postApi,
		CategoryApi: categoryApi,
		AdminApi:    adminApi,
		AuthorApi:   authorApi,
	}
}
