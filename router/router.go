package router

import (
	"strings"

	"github.com/kataras/iris/core/router"
	"github.com/kataras/iris/v12"
	"github.com/spf13/viper"
)

type IFnRegisterRoute = func(rgPublic router.Party, rgAuth router.Party)

var (
	gfnRouters []IFnRegisterRoute
)

const (
	EnvDev byte = 0x01 // 0001
	EnvPro byte = 0x02 // 0010
)

// bit(0) means development, bit(1) means production
var (
	environment byte
	isDev       func() bool
	isPro       func() bool
)

func preInit() {
	iEnv := strings.ToLower(viper.GetString("env"))
	if iEnv == "dev" || iEnv == "develop" {
		environment = EnvDev
	} else if iEnv == "pro" || iEnv == "production" {
		environment = EnvPro
	} else {
		environment = EnvDev
	}

	isDev = func() bool {
		return environment&EnvDev == EnvDev
	}
	isPro = func() bool {
		return environment&EnvPro == EnvPro
	}

}

func RegisterRoute(fn IFnRegisterRoute) {
	if fn == nil {
		return
	}
	gfnRouters = append(gfnRouters, fn)
}

func InitRoute() {
	app := iris.New()

	if isDev() {
		app.Logger().SetLevel("debug")
	}
	if isPro() {
		app.Logger().SetLevel("info")
	}
}
