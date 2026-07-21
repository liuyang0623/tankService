package appconfig

import (
	"encoding/json"
	"testing"
)

// TestAppConfig_TableName 验证表名固定为 app_config。
func TestAppConfig_TableName(t *testing.T) {
	if (AppConfig{}).TableName() != "app_config" {
		t.Errorf("expected table name app_config, got %s", (AppConfig{}).TableName())
	}
}

// TestAppConfigResponse_JSONTag 验证 DTO 序列化使用 camelCase auditMode。
func TestAppConfigResponse_JSONTag(t *testing.T) {
	b, err := json.Marshal(AppConfigResponse{AuditMode: true})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	got := string(b)
	want := `{"auditMode":true}`
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}
