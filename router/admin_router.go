package router

import (
	"github.com/kataras/iris/v12"
)

func RegisterAdminRoutes(rgPublic iris.Party, rgPrivate iris.Party, apiContainer *AppContainer) {
	adminApi := apiContainer.AdminApi

	{
		rgPublic.Post("/login", adminApi.Login)
	}
}
