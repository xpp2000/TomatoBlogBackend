package global

// AppConfig 对应整个 settings.yml
type AppConfig struct {
	// ... 你现有的 App, Db, Server 等结构体嵌套
	App      AppConfigBlock      `mapstructure:"app"`
	Mode     ModeConfigBlock     `mapstructure:"mode"`
	Db       DbConfigBlock       `mapstructure:"db"`
	Server   ServerConfigBlock   `mapstructure:"server"`
	Jwt      JwtConfigBlock      `mapstructure:"jwt"`
	Log      LogConfig           `mapstructure:"log"` // 昨天抽离的日志配置
	Cdn      CdnConfigBlock      `mapstructure:"cdn"`
	Modules  ModulesConfigBlock  `mapstructure:"modules"`
	DataInit DataInitConfigBlock `mapstructure:"dataInit"`

	Oss OssConfig `mapstructure:"oss"` // ⭐️ 映射 oss 节点
}

// LogConfig 日志配置映射
type LogConfig struct {
	MaxSize    int `mapstructure:"MaxSize"`
	MaxBackups int `mapstructure:"MaxBackups"`
	MaxAge     int `mapstructure:"MaxAge"`
}

// OssConfig 专门用于存储配置
type OssConfig struct {
	Provider     string `mapstructure:"provider"`
	Endpoint     string `mapstructure:"endpoint"`
	AccessKey    string `mapstructure:"accessKey"`
	SecretKey    string `mapstructure:"secretKey"`
	BucketName   string `mapstructure:"bucketName"`
	UseSSL       bool   `mapstructure:"useSSL"`
	CustomDomain string `mapstructure:"customDomain"`
}
type AppConfigBlock struct {
	Name string `mapstructure:"name"`
}

type ModeConfigBlock struct {
	Develop bool `mapstructure:"develop"`
}

type DbConfigBlock struct {
	Dsn         string `mapstructure:"dsn"`
	MaxIdleConn int    `mapstructure:"maxIdleConn"`
	MaxOpenConn int    `mapstructure:"maxOpenConn"`
}
type ServerConfigBlock struct {
	Port int `mapstructure:"port"`
}

type JwtConfigBlock struct {
	Xx          int    `mapstructure:"xx"`          // 可能是随手敲的测试残留？建议删除
	TokenExpire int    `mapstructure:"tokenExpire"` // token有效时长(分钟)
	SigningKey  string `mapstructure:"signingKey"`  // 签名使用的key
}

type CdnConfigBlock struct {
	DevCDN string `mapstructure:"devCDN"`
}

// ModulesConfigBlock 模块开关配置 (注意这里的嵌套处理)
type ModulesConfigBlock struct {
	Post PostModuleConfig `mapstructure:"post"`
}

// PostModuleConfig 文章模块的具体配置
type PostModuleConfig struct {
	Enable bool `mapstructure:"enable"`
}

type DataInitConfigBlock struct {
	InitAdminName  string `mapstructure:"InitAdminName"`
	InitAdminPsw   string `mapstructure:"InitAdminPsw"`
	InitAuthorName string `mapstructure:"InitAuthorName"`
	InitAuthorPsw  string `mapstructure:"InitAuthorPsw"`
}
