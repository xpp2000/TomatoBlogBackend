package cmd

import (
	"tomatoBlogDB/conf"
	"tomatoBlogDB/global"
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
	global.DB, err := conf.InitDB()
	if err != nil {
		initErr = utils.AppendError()
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
