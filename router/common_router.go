package router

import "github.com/kataras/iris/v12"

func RegisterCommonRoutes(rgPublic iris.Party, rgPrivate iris.Party, apiContainer *AppContainer) {
	commonApi := apiContainer.CommonApi

	commonParty := rgPrivate.Party("/common")
	{
		commonParty.Post("/uploadimg", iris.LimitRequestBodySize(5*iris.MB), commonApi.UploadImage)
	}
}
