package router

import (
	"github.com/kataras/iris/v12"
)

func RegisterPostRoutes(rgPublic iris.Party, rgPrivate iris.Party, apiContainer *AppContainer) {
	postApi := apiContainer.PostApi
	cateApi := apiContainer.CategoryApi
	p := rgPrivate.Party("/post")
	{
		p.Post("/", postApi.AddPost)
		p.Put("/{id}", postApi.UpdatePost)
		p.Patch("/{id}/status", postApi.UpdateStatus)
		p.Delete("/{id}", postApi.DeletePost)
	}

	{
		rgPublic.Get("/posts", postApi.ListPosts)
		rgPublic.Get("/post/{slug_or_id}", postApi.GetPost)

		rgPublic.Get("/category/{name_or_id}", cateApi.GetCategory)
		rgPublic.Get("/categories", cateApi.ListCategory)
	}

}
