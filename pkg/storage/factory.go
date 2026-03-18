package storage

import (
	"fmt"
	"tomatoBlogDB/global"
)

// InitStorage 根据 yaml 配置文件，动态挂载上传驱动
func InitStorage(cfg global.OssConfig) error {
	var err error

	switch cfg.Provider {
	case "minio", "r2", "aliyun":
		// 🌟 奇迹发生的地方：因为它们都兼容 S3 协议，所以我们复用同一个 Driver！
		GlobalUploader, err = NewS3Driver(
			cfg.Endpoint,
			cfg.AccessKey,
			cfg.SecretKey,
			cfg.BucketName,
			cfg.CustomDomain,
			cfg.UseSSL,
		)
	// case "qiniu":
	// 如果未来你用了完全不兼容 S3 的七牛云，你可以写一个 NewQiniuDriver
	// GlobalUploader = NewQiniuDriver(...)
	default:
		return fmt.Errorf("Doesn't support OSS provider: %s", cfg.Provider)
	}

	if err != nil {
		return fmt.Errorf("fail to init OSS service: %w", err)
	}

	return nil
}
