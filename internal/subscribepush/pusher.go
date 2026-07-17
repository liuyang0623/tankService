// Package subscribepush 连接关注事件与微信订阅消息推送：查配额/openid、
// 按配额推送、成功后扣减配额。作为 follow 的 subscribePusher 注入实现。
package subscribepush

import (
	"context"
	"log"
	"time"
)

// sender 抽象微信订阅消息发送（由 internal/wechat.Client 满足）。
type sender interface {
	SendSubscribeMessage(ctx context.Context, openid, tplID string, data map[string]any) error
}

// store 抽象订阅目标的数据查询（openid、配额、昵称）与配额扣减。
type store interface {
	GetSubscribeTarget(ctx context.Context, userID uint) (openid string, quota int, err error)
	DecrSubscribeQuota(ctx context.Context, userID uint) error
	GetNickname(ctx context.Context, userID uint) (string, error)
}

// Pusher 实现 follow 的 subscribePusher 接口。
type Pusher struct {
	sender sender
	store  store
	tplID  string
}

// New 创建 Pusher。tplID 为微信订阅消息模板 ID。
func New(s sender, st store, tplID string) *Pusher {
	return &Pusher{sender: s, store: st, tplID: tplID}
}

// PushFollow 异步推送关注订阅消息，fire-and-forget，不阻塞调用方。
func (p *Pusher) PushFollow(ctx context.Context, targetID, actorID uint) {
	go func() {
		c, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.pushFollowSync(c, targetID, actorID)
	}()
}

// pushFollowSync 同步核心：查目标配额/openid → 有配额且 openid 非空则发送 →
// 成功后扣减配额。任何失败只记日志。
func (p *Pusher) pushFollowSync(ctx context.Context, targetID, actorID uint) {
	openid, quota, err := p.store.GetSubscribeTarget(ctx, targetID)
	if err != nil {
		log.Printf("subscribe push: lookup target %d failed: %v", targetID, err)
		return
	}
	if quota <= 0 || openid == "" {
		return // 无配额或无 openid，不推送（站内通知已兜底）
	}

	nickname, _ := p.store.GetNickname(ctx, actorID)
	if nickname == "" {
		nickname = "有人"
	}
	data := map[string]any{
		"thing1": map[string]string{"value": nickname},
		"time2":  map[string]string{"value": time.Now().Format("2006-01-02 15:04")},
	}

	if err := p.sender.SendSubscribeMessage(ctx, openid, p.tplID, data); err != nil {
		log.Printf("subscribe push: send to %d failed: %v", targetID, err)
		return
	}
	if err := p.store.DecrSubscribeQuota(ctx, targetID); err != nil {
		log.Printf("subscribe push: decr quota for %d failed: %v", targetID, err)
	}
}
