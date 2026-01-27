package conf

import (
	"time"
	"tomatoBlogDB/global"
	"tomatoBlogDB/model"

	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func initAdmin(db *gorm.DB) {
	var admin model.Admin

	// = check whether default admin exist
	err := db.Where("name=?", "admin").First(&admin).Error

	// if exist return
	if err == nil {
		global.Logger.Info("Default admin has existed")
		return
	}

	if err == gorm.ErrRecordNotFound {
		global.Logger.Info("Default admin doesn't exist, creating admin......")

		newAdmin := model.Admin{
			Name:     "tomato",
			RealName: "Default Admin",
		}

		if err := newAdmin.SetPassword("asd123"); err != nil {
			global.Logger.Panic("fail to encrypt password", err)
		}

		if err := db.Create(&newAdmin).Error; err != nil {
			global.Logger.Panic("fail to create default admin", err)
		}
	}
	global.Logger.Info("create admin successfully")
}

func InitDB() (*gorm.DB, error) {
	logMode := logger.Info

	if !viper.GetBool("mode.develop") {
		logMode = logger.Error
	}

	db, err := gorm.Open(postgres.Open(viper.GetString("db.dsn")), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "sys_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logMode),
	})

	if err != nil {
		return nil, err
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(viper.GetInt("db.maxIdleConn"))
	sqlDB.SetMaxOpenConns(viper.GetInt("db.maxOpenConn"))
	sqlDB.SetConnMaxLifetime(time.Hour)

	// = migrate struct
	db.AutoMigrate(&model.Admin{})

	// = seed data
	initAdmin(db)

	return db, nil

}
