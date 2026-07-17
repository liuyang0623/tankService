package subscribepush

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// GormStore 是 store 接口的 gorm 实现，直接查询 users 表。
type GormStore struct {
	db *gorm.DB
}

// NewGormStore 创建基于 *gorm.DB 的 store。
func NewGormStore(db *gorm.DB) *GormStore {
	return &GormStore{db: db}
}

// GetSubscribeTarget 返回目标用户的 openid 与剩余订阅配额。
func (g *GormStore) GetSubscribeTarget(ctx context.Context, userID uint) (string, int, error) {
	var row struct {
		Openid               string
		SubscribeFollowQuota int
	}
	err := g.db.WithContext(ctx).
		Table("users").
		Select("openid", "subscribe_follow_quota").
		Where("id = ?", userID).
		Scan(&row).Error
	if err != nil {
		return "", 0, fmt.Errorf("get subscribe target: %w", err)
	}
	return row.Openid, row.SubscribeFollowQuota, nil
}

// DecrSubscribeQuota 原子扣减配额（不低于 0）。
func (g *GormStore) DecrSubscribeQuota(ctx context.Context, userID uint) error {
	err := g.db.WithContext(ctx).
		Table("users").
		Where("id = ? AND subscribe_follow_quota > 0", userID).
		UpdateColumn("subscribe_follow_quota", gorm.Expr("subscribe_follow_quota - 1")).Error
	if err != nil {
		return fmt.Errorf("decr subscribe quota: %w", err)
	}
	return nil
}

// GetNickname 返回用户昵称。
func (g *GormStore) GetNickname(ctx context.Context, userID uint) (string, error) {
	var nickname string
	err := g.db.WithContext(ctx).
		Table("users").
		Select("nickname").
		Where("id = ?", userID).
		Scan(&nickname).Error
	if err != nil {
		return "", fmt.Errorf("get nickname: %w", err)
	}
	return nickname, nil
}
