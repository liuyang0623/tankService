package appconfig

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// AppConfigService 处理全局应用配置业务逻辑。
type AppConfigService struct {
	db *gorm.DB
}

// NewAppConfigService 构造。
func NewAppConfigService(db *gorm.DB) *AppConfigService {
	return &AppConfigService{db: db}
}

// GetConfig 读取单行应用配置。
// 无记录时返回种子默认（auditMode=false，正常模式），不视为错误——
// 保证接口在表尚未 seed 时仍可用。
func (s *AppConfigService) GetConfig(ctx context.Context) (*AppConfigResponse, error) {
	var cfg AppConfig
	err := s.db.WithContext(ctx).First(&cfg, singletonID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &AppConfigResponse{AuditMode: false}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get app config: %w", err)
	}
	return &AppConfigResponse{AuditMode: cfg.AuditMode}, nil
}

// Seed 幂等地插入种子配置行（audit_mode=false）。
// 仅当单行不存在时插入，已存在则不覆盖运营已切换的值。
func (s *AppConfigService) Seed(ctx context.Context) error {
	var count int64
	if err := s.db.WithContext(ctx).Model(&AppConfig{}).Where("id = ?", singletonID).Count(&count).Error; err != nil {
		return fmt.Errorf("count app config: %w", err)
	}
	if count > 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Create(&AppConfig{ID: singletonID, AuditMode: false}).Error; err != nil {
		return fmt.Errorf("seed app config: %w", err)
	}
	return nil
}
