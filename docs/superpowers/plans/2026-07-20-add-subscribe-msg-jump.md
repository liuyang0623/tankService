---
change: add-subscribe-msg-jump
design-doc: docs/superpowers/specs/2026-07-20-add-subscribe-msg-jump-design.md
base-ref: 83f285da1f3d4645b2ef626e4bcd36f2c2794043
---

# 实施计划：关注订阅消息跳转（仅后端）

引用设计文档：`docs/superpowers/specs/2026-07-20-add-subscribe-msg-jump-design.md`

## 任务 1：扩展 `SendSubscribeMessage` 支持 page

- 修改 `internal/wechat/client.go`：`SendSubscribeMessage` 签名增加 `page string`（置于 `data` 之后），请求 body 仅当 `page != ""` 时写入 `page` 字段。
- 更新 `internal/subscribepush/pusher.go` 的 `sender` 接口方法签名同步加 `page string`。
- TDD：先在 `internal/wechat/client_test.go` 补失败测试（page 非空写入 / 空不写入）。

## 任务 2：Pusher 下发关注通知携带跳转页

- `internal/subscribepush/pusher.go` 定义常量 `notificationPage = "pages/notifications/index"`。
- `pushFollowSync` 调用 sender 时传入该常量。
- TDD：在 `internal/subscribepush/pusher_test.go` 断言传入的 page 值，并覆盖配额/openid 现有行为不回归。

## 任务 3：验证

- `go build ./...`
- `go test ./internal/wechat/... ./internal/subscribepush/...`
