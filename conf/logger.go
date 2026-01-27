package conf

import (
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger() *zap.SugaredLogger {
	logMode := zapcore.DebugLevel
	if !viper.GetBool("mode.develop") {
		logMode = zapcore.InfoLevel
	}
	// core := zapcore.NewCore(getEncoder(), getWriteSyncer(), logMode)
	core := zapcore.NewCore(getEncoder(), zapcore.NewMultiWriteSyncer(getWriteSyncer(), zapcore.AddSync(os.Stdout)), logMode)
	return zap.New(core).Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString(t.Local().Format(time.DateTime))
	}
	return zapcore.NewJSONEncoder(encoderConfig)
}

// set the log path according to system.
func getWriteSyncer() zapcore.WriteSyncer {
	fileName := "TomatoBlogManager.log"

	// file directory
	var logDir string
	if viper.GetBool("mode.develop") || runtime.GOOS == "windows" {
		// --- Windows ---
		rootDir, _ := os.Getwd()
		logDir = filepath.Join(rootDir, "log")
	} else {
		// --- Linux ---
		// usually /var/log/<appName>
		appName := viper.GetString("app.name")
		if appName == "" {
			appName = "TomatoBlogCMS"
		}
		logDir = filepath.Join("/var/log", appName)

	}

	// check existence.
	// if logDir has existed, os.MkdirAll() will return nil safely
	_ = os.MkdirAll(logDir, 0755)

	// full path
	logFilePath := filepath.Join(logDir, fileName)

	lumberjackSyncer := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    viper.GetInt("log.MaxSize"), //megabytes
		MaxBackups: viper.GetInt("log.MaxBackups"),
		MaxAge:     viper.GetInt("log.MaxAge"),
		Compress:   true,
		LocalTime:  true,
	}
	return zapcore.AddSync(lumberjackSyncer)

}
