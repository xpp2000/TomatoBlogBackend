package conf

import (
	"fmt"
	"tomatoBlogDB/global"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.SetConfigName("settings")
	viper.SetConfigType("yml")
	viper.AddConfigPath("./conf/")
	// 2.
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Load Config Error: %s", err.Error()))
	}
	// 3.
	if err := viper.Unmarshal(&global.Config); err != nil {
		panic(fmt.Errorf("Unmarshal Error: %w \n", err))
	}
	fmt.Println(" Viper config loaded successfully: running env =", global.Config.Mode.Develop)
}
