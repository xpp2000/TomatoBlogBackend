package cmd

import (
	"tomatoBlogDB/conf"
	"tomatoBlogDB/global"
	"tomatoBlogDB/router"
	"tomatoBlogDB/utils"
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
	global.DB = db
	if err != nil {
		initErr = utils.AppendError(initErr, err)
	}

	if initErr != nil {
		if global.Logger != nil {
			global.Logger.Error(initErr.Error())
		}
		panic(initErr.Error())
	}

	// = init router
	router.InitRouter()
}
