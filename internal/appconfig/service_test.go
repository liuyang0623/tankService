package appconfig

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// newTestDB 用纯 Go sqlite 内存库建 app_config 表，供 service 落库测试。
func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&AppConfig{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// TestUpdateAuditMode_PersistsAndReadsBack 验证更新后 GetConfig 返回新值。
func TestUpdateAuditMode_PersistsAndReadsBack(t *testing.T) {
	svc := NewAppConfigService(newTestDB(t))
	ctx := context.Background()

	// 初始无记录：GetConfig 返回默认 false
	got, err := svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("initial GetConfig: %v", err)
	}
	if got.AuditMode {
		t.Fatalf("expected initial auditMode=false, got true")
	}

	// 更新为 true（表未 seed，走 upsert 创建单行）
	if err := svc.UpdateAuditMode(ctx, true); err != nil {
		t.Fatalf("UpdateAuditMode(true): %v", err)
	}
	got, err = svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig after set true: %v", err)
	}
	if !got.AuditMode {
		t.Errorf("expected auditMode=true after update, got false")
	}

	// 再切回 false，确认可反复更新同一行
	if err := svc.UpdateAuditMode(ctx, false); err != nil {
		t.Fatalf("UpdateAuditMode(false): %v", err)
	}
	got, err = svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("GetConfig after set false: %v", err)
	}
	if got.AuditMode {
		t.Errorf("expected auditMode=false after second update, got true")
	}

	// 确认全表始终只有单行（upsert 未产生多行）
	var count int64
	if err := svc.db.Model(&AppConfig{}).Count(&count).Error; err != nil {
		t.Fatalf("count rows: %v", err)
	}
	if count != 1 {
		t.Errorf("expected exactly 1 row, got %d", count)
	}
}
