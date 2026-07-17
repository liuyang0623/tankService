package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	Port          string
	APIPrefix     string
	CORSOrigin    string
	JWTSecret     string
	WechatAppID   string
	WechatSecret  string
	WechatSubscribeTplFollow string
	UpyunBucket   string
	UpyunOperator string
	UpyunPassword string
	UpyunEndpoint string
	UpyunDomain   string
}

func Load() (*Config, error) {
	// 加载 .env 文件中的环境变量（已存在的环境变量不会被覆盖）
	_ = godotenv.Load(".env")

	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		Port:          getEnvOrDefault("PORT", "3000"),
		APIPrefix:     getEnvOrDefault("API_PREFIX", "api/v1"),
		CORSOrigin:    getEnvOrDefault("CORS_ORIGIN", "*"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		WechatAppID:   os.Getenv("WECHAT_APPID"),
		WechatSecret:  os.Getenv("WECHAT_SECRET"),
		WechatSubscribeTplFollow: os.Getenv("WECHAT_SUBSCRIBE_TPL_FOLLOW"),
		UpyunBucket:   os.Getenv("UPYUN_BUCKET"),
		UpyunOperator: getEnvOrDefault("UPYUN_OPERATOR", "liuyang"),
		UpyunPassword: os.Getenv("UPYUN_PASSWORD"),
		UpyunEndpoint: getEnvOrDefault("UPYUN_ENDPOINT", "v0.api.upyun.com"),
		UpyunDomain:   os.Getenv("UPYUN_DOMAIN"),
	}
	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
