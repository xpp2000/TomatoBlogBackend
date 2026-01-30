package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"tomatoBlogDB/conf"
	"tomatoBlogDB/global"
	"tomatoBlogDB/router"
	"tomatoBlogDB/utils"

	"github.com/kataras/iris/v12"
	"github.com/spf13/viper"
)

func Clean() {

}

// 1. init global settings
// 2. init Log
// 3. init DB

func Start() {
	var initErr error

	// 1.
	conf.InitConfig()

	// 2.
	global.Logger = conf.InitLogger()

	// 3.
	db, err := conf.InitDB()
	if err != nil {
		initErr = utils.AppendError(initErr, err)
	}

	if initErr != nil {
		if global.Logger != nil {
			global.Logger.Error(initErr.Error())
		}
		panic(initErr.Error())
	}
	global.DB = db

	// = init router
	app := router.NewApp()

	port := viper.GetString("server.port")
	if port == "" {
		port = "8888"
	}
	addr := fmt.Sprintf(":%s", port)

	// 4.graceful shutdown
	go func() {
		global.Logger.Info("Server is running at " + addr)
		if err := app.Listen(addr, iris.WithoutInterruptHandler); err != nil && err != http.ErrServerClosed {
			global.Logger.Error("Server Start Failed: " + err.Error())
		}
	}()

	// 5. listen to interruption signal
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// 阻塞在这里，直到收到信号
	<-quit
	global.Logger.Info("Shutdown Server ...")

	// 6.shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		global.Logger.Error("Server Shutdown Error: " + err.Error())
	}

}
