package router

import (
	stdContext "context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	_ "tomatoBlogDB/docs"
	"tomatoBlogDB/global"
	"tomatoBlogDB/middleware"

	"github.com/iris-contrib/middleware/cors"
	"github.com/iris-contrib/swagger/v12"
	"github.com/iris-contrib/swagger/v12/swaggerFiles"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/core/router"
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

func setEnvFunc() {
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

func NewApp() *iris.Application {
	setEnvFunc()

	app := iris.New()

	// +. 配置 CORS 规则
	crs := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "PATCH"},
	})
	// +. 挂载 CORS 中间件
	// ⚠️ 极其重要：必须使用 UseRouter 而不是 Use！
	// 因为 OPTIONS 预检请求需要在路由匹配之前就被拦截并响应。
	app.UseRouter(crs)

	// 1. Log level
	if isDev() {
		app.Logger().SetLevel("debug")
	} else {
		app.Logger().SetLevel("info")
	}

	// 2. register middleware
	app.Use(iris.Compression)

	// 3. router group
	v1 := app.Party("/api/v1")

	publicGroup := v1.Party("/")
	privateGroup := v1.Party("/private")
	privateGroup.Use(middleware.Auth())

	// ==== register module ====
	apiContainer := InitDI(global.DB)
	// ---- compulsory module ----
	RegisterAdminRoutes(publicGroup, privateGroup, apiContainer)
	// ---- selective modules ----
	if viper.GetBool("modules.post.enable") {
		RegisterPostRoutes(publicGroup, privateGroup, apiContainer)
	}

	// 4. Swagger ()
	if isDev() {
		swaggerConfig := &swagger.Config{
			URL: fmt.Sprintf("http://localhost:%s/swagger/doc.json", viper.GetString("server.port")),
		}
		app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(swaggerConfig, swaggerFiles.Handler))
	}
	return app
}

// = abolished
func InitRouter() {
	setEnvFunc()
	app := iris.New()

	if isDev() {
		app.Logger().SetLevel("debug")
	}
	if isPro() {
		app.Logger().SetLevel("info")
	}

	InitBasePlatformRoutes()

	rgApi := app.Party("/")
	rgAdmin := app.Party("/admin")

	// 全局2MB限制，对于Upload再单独放宽
	rgApi.Use(iris.LimitRequestBodySize(2 << 20))

	rgAdmin.Use(middleware.Auth())

	// = register module routers
	for _, fnRegisterRoute := range gfnRouters {
		fnRegisterRoute(rgApi, rgAdmin)
	}

	// = integrate swagger
	if isDev() {
		config := &swagger.Config{
			URL:         "http://localhost:8888/swagger/doc.json",
			DeepLinking: true,
		}
		app.Get("/swagger/{any:path}", swagger.CustomWrapHandler(config, swaggerFiles.Handler))
	}

	// = reading server settings from viper
	stPort := viper.GetString("server.port")
	if stPort == "" {
		stPort = "8888"
	}

	// graceful shutdown
	idleConnsClosed := make(chan struct{})
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch,
			// kill -SIGINT XXXX or Ctrl+c
			os.Interrupt,
			syscall.SIGINT, // register that too, it should be ok
			// os.Kill  is equivalent with the syscall.Kill
			os.Kill,
			syscall.SIGKILL, // register that too, it should be ok
			// kill -SIGTERM XXXX
			syscall.SIGTERM,
		)
		select {
		case <-ch:
			println("shutdown...")

			timeout := 10 * time.Second
			ctx, cancel := stdContext.WithTimeout(stdContext.Background(), timeout)
			defer cancel()
			app.Shutdown(ctx)
			close(idleConnsClosed)
		}
	}()

	//= Start the server
	err := app.Run(
		iris.Addr("localhost:8888"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
		iris.WithoutInterruptHandler,
	)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("Start Server Error: %s", err.Error()))
	}
	global.Logger.Info("Start TomatoBlogDBServer successfully")

	<-idleConnsClosed
	global.Logger.Info("Stop Server successfully")
}

func InitBasePlatformRoutes() {

}
