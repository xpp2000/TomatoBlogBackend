package router

import (
	"github.com/kataras/iris/v12"
)

func RegisterAdminRoutes(rgPublic iris.Party, rgPrivate iris.Party, apiContainer *AppContainer) {
	adminApi := apiContainer.AdminApi
	p := rgPrivate.Party("/admin")

	{
		p.Post("/", adminApi.AddAdmin)
		p.Patch("/{id}/status", adminApi.UpdateStatus)
		p.Delete("/{id}", adminApi.DeleteAdmin)
	}
	rgPrivate.Get("/admins", adminApi.ListAdmin)
	{
		rgPublic.Post("/login", adminApi.Login)
	}

}
