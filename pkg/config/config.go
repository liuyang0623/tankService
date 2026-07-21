package config

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL              string
	Port                     string
	APIPrefix                string
	CORSOrigin               string
	JWTSecret                string
	WechatAppID              string
	WechatSecret             string
	WechatSubscribeTplFollow string
	UpyunBucket              string
	UpyunOperator            string
	UpyunPassword            string
	UpyunEndpoint            string
	UpyunDomain              string
	// AdminOpenids 管理员白名单（openid 列表），从 ADMIN_OPENIDS（逗号分隔）解析。
	// 空表示无人拥有管理员权限（fail-closed）。
	AdminOpenids []string
}

func Load() (*Config, error) {
	// 加载 .env 文件中的环境变量（已存在的环境变量不会被覆盖）
	_ = godotenv.Load(".env")

	cfg := &Config{
		DatabaseURL:              os.Getenv("DATABASE_URL"),
		Port:                     getEnvOrDefault("PORT", "3000"),
		APIPrefix:                getEnvOrDefault("API_PREFIX", "api/v1"),
		CORSOrigin:               getEnvOrDefault("CORS_ORIGIN", "*"),
		JWTSecret:                os.Getenv("JWT_SECRET"),
		WechatAppID:              os.Getenv("WECHAT_APPID"),
		WechatSecret:             os.Getenv("WECHAT_SECRET"),
		WechatSubscribeTplFollow: os.Getenv("WECHAT_SUBSCRIBE_TPL_FOLLOW"),
		UpyunBucket:              os.Getenv("UPYUN_BUCKET"),
		UpyunOperator:            getEnvOrDefault("UPYUN_OPERATOR", "liuyang"),
		UpyunPassword:            os.Getenv("UPYUN_PASSWORD"),
		UpyunEndpoint:            getEnvOrDefault("UPYUN_ENDPOINT", "v0.api.upyun.com"),
		UpyunDomain:              os.Getenv("UPYUN_DOMAIN"),
		AdminOpenids:             parseCommaList(os.Getenv("ADMIN_OPENIDS")),
	}
	return cfg, nil
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// parseCommaList 将逗号分隔字符串解析为去空白、去空项的切片。
// 空输入返回 nil，保证白名单未配置时为空（fail-closed）。
func parseCommaList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			result = append(result, v)
		}
	}
	return result
}
