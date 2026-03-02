package router

import (
	"github.com/kataras/iris/v12"
)

func RegisterPostRoutes(rgPublic iris.Party, rgPrivate iris.Party, apiContainer *AppContainer) {
	postApi := apiContainer.PostApi
	cateApi := apiContainer.CategoryApi
	authorApi := apiContainer.AuthorApi
	p := rgPrivate.Party("/post")
	{
		p.Post("/", postApi.AddPost)
		p.Put("/{id}", postApi.UpdatePost)
		p.Patch("/{id}/status", postApi.UpdateStatus)
		p.Delete("/{id}", postApi.DeletePost)
	}
	pCategory := rgPrivate.Party("/category")
	{
		pCategory.Post("/", cateApi.AddCategory)
		pCategory.Delete("/{id}", cateApi.DeleteCategory)
	}
	pAuthor := rgPrivate.Party("/author")
	{
		// pAuthor.Post("/", authorApi.AddAuthor)
		pAuthor.Delete("/{id}", authorApi.DeleteAuthor)
	}

	{
		rgPublic.Get("/posts", postApi.ListPosts)
		rgPublic.Get("/post/{slug_or_id}", postApi.GetPost)

		rgPublic.Get("/category/{name_or_id}", cateApi.GetCategory)
		rgPublic.Get("/categories", cateApi.ListCategory)

		rgPublic.Get("/author/{name_or_id}", authorApi.GetAuthor)
		rgPublic.Get("/authors", authorApi.ListAuthors)

	}

}
