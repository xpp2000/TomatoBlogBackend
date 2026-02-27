package router

import (
	"tomatoBlogDB/api"

	"github.com/kataras/iris/v12"
)

func RegisterAdminRoutes(rgPublic iris.Party, rgPrivate iris.Party) {
	adminApi := api.NewAdminApi()

	{
		rgPublic.Post("/login", adminApi.Login)
	}
}
