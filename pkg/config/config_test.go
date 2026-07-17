package config

import (
	"os"
	"testing"
)

// clearEnv 清除所有配置相关的环境变量，确保测试隔离
func clearEnv() {
	envVars := []string{
		"DATABASE_URL", "PORT", "API_PREFIX", "CORS_ORIGIN",
		"JWT_SECRET", "WECHAT_APPID", "WECHAT_SECRET",
		"UPYUN_BUCKET", "UPYUN_OPERATOR", "UPYUN_PASSWORD",
		"UPYUN_ENDPOINT", "UPYUN_DOMAIN",
	}
	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

// TestLoad_Defaults 测试在没有任何环境变量时，默认值是否正确
func TestLoad_Defaults(t *testing.T) {
	clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Port default", cfg.Port, "3000"},
		{"APIPrefix default", cfg.APIPrefix, "api/v1"},
		{"CORSOrigin default", cfg.CORSOrigin, "*"},
		{"UpyunOperator default", cfg.UpyunOperator, "liuyang"},
		{"UpyunEndpoint default", cfg.UpyunEndpoint, "v0.api.upyun.com"},
		{"DatabaseURL empty", cfg.DatabaseURL, ""},
		{"JWTSecret empty", cfg.JWTSecret, ""},
		{"WechatAppID empty", cfg.WechatAppID, ""},
		{"WechatSecret empty", cfg.WechatSecret, ""},
		{"UpyunBucket empty", cfg.UpyunBucket, ""},
		{"UpyunPassword empty", cfg.UpyunPassword, ""},
		{"UpyunDomain empty", cfg.UpyunDomain, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

// TestLoad_EnvOverride 测试设置环境变量时是否正确覆盖默认值
func TestLoad_EnvOverride(t *testing.T) {
	clearEnv()

	// 设置测试用的环境变量
	os.Setenv("DATABASE_URL", "mysql://user:pass@localhost:3306/testdb")
	os.Setenv("PORT", "8080")
	os.Setenv("API_PREFIX", "api/v2")
	os.Setenv("CORS_ORIGIN", "https://example.com")
	os.Setenv("JWT_SECRET", "supersecret")
	os.Setenv("WECHAT_APPID", "wx123456")
	os.Setenv("WECHAT_SECRET", "wechatsecret")
	os.Setenv("UPYUN_BUCKET", "mybucket")
	os.Setenv("UPYUN_OPERATOR", "myoperator")
	os.Setenv("UPYUN_PASSWORD", "mypassword")
	os.Setenv("UPYUN_ENDPOINT", "v1.api.upyun.com")
	os.Setenv("UPYUN_DOMAIN", "https://cdn.example.com")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned unexpected error: %v", err)
	}

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"DatabaseURL override", cfg.DatabaseURL, "mysql://user:pass@localhost:3306/testdb"},
		{"Port override", cfg.Port, "8080"},
		{"APIPrefix override", cfg.APIPrefix, "api/v2"},
		{"CORSOrigin override", cfg.CORSOrigin, "https://example.com"},
		{"JWTSecret override", cfg.JWTSecret, "supersecret"},
		{"WechatAppID override", cfg.WechatAppID, "wx123456"},
		{"WechatSecret override", cfg.WechatSecret, "wechatsecret"},
		{"UpyunBucket override", cfg.UpyunBucket, "mybucket"},
		{"UpyunOperator override", cfg.UpyunOperator, "myoperator"},
		{"UpyunPassword override", cfg.UpyunPassword, "mypassword"},
		{"UpyunEndpoint override", cfg.UpyunEndpoint, "v1.api.upyun.com"},
		{"UpyunDomain override", cfg.UpyunDomain, "https://cdn.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

// TestLoad_ReturnNoError 测试 Load 函数不返回错误
func TestLoad_ReturnNoError(t *testing.T) {
	clearEnv()

	_, err := Load()
	if err != nil {
		t.Errorf("Load() should not return error, got: %v", err)
	}
}
