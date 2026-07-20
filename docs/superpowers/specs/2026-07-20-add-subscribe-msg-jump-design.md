---
comet_change: add-subscribe-msg-jump
role: technical-design
canonical_spec: openspec
archived-with: 2026-07-20-add-subscribe-msg-jump
status: final
---

# 关注订阅消息跳转 — 技术设计

## 背景

关注订阅消息下发链路：`follow` 触发关注事件 → `subscribepush.Pusher.pushFollowSync` 查配额/openid → `wechat.Client.SendSubscribeMessage(ctx, openid, tplID, data)` 下发。当前请求 body 只含 `touser`、`template_id`、`data`，未携带微信 `subscribe/send` 支持的 `page` 参数，用户点击订阅消息只能打开小程序默认首页。

前端小程序已存在系统通知页 `pages/notifications/index`（`app.config.ts` 注册，标题「系统通知」）。该页进入时调用 `GET /notifications` 拉列表 + 整体标记已读，自给自足，深链进入也能渲染，不依赖任何单条通知详情接口。

## 目标 / 非目标

**目标：**
- 关注订阅消息携带 `page`，用户点击后直达 `pages/notifications/index`。
- `SendSubscribeMessage` 具备透传任意 `page` 的能力，供后续其他订阅消息复用。

**非目标：**
- 不新增单条通知详情接口（前端列表页自给自足）。
- 不改通知模型、通知类型、配额/授权逻辑。
- 不动前端代码。

## 详细设计

### 改动点 1：`internal/wechat/client.go` — `SendSubscribeMessage`

签名扩展：

```go
func (c *Client) SendSubscribeMessage(ctx context.Context, openid, tplID string, data map[string]any, page string) error
```

请求 body 构造：在原有 `touser`/`template_id`/`data` 基础上，仅当 `page != ""` 时追加 `body["page"] = page`。空串不写入该字段，保证向后兼容——微信 `page` 是可选字段，缺省即不跳转。

### 改动点 2：`internal/subscribepush/pusher.go`

- 包内定义常量 `const notificationPage = "pages/notifications/index"`。
- `sender` 接口的 `SendSubscribeMessage` 方法签名同步增加 `page string` 形参。
- `pushFollowSync` 调用 sender 时传入 `notificationPage`。

### 数据流

```
follow 关注事件
  → Pusher.pushFollowSync(target, actor)
      查配额/openid → 拼装 data 字段
      → sender.SendSubscribeMessage(ctx, openid, tplID, data, "pages/notifications/index")
          → body 含 page 字段 → 微信 subscribe/send
  → 用户点击订阅消息 → 打开小程序 pages/notifications/index
      → loadList() + markRead() → GET /notifications → 渲染系统通知
```

## 关键取舍与风险

- **取舍**：直接加 `page string` 形参而非 options 封装（YAGNI，仅一个字段）。
- **风险1（路径漂移）**：前端若重命名通知页路由，后端常量 `notificationPage` 需同步更新。契约路径已在此固化。
- **风险2（微信 page 校验）**：`page` 必须为小程序已发布合法页面路径；`pages/notifications/index` 已注册，风险低。page 空值不写入字段，不会因此导致下发失败。
- **影响面**：`internal/wechat`（client + 接口调用点）、`internal/subscribepush`（sender 接口 + pusher）及各自测试。

## 测试策略

- `internal/wechat/client_test.go`：断言 `page` 非空时请求 body 含 `page` 字段、为空时不含该字段。
- `internal/subscribepush/pusher_test.go`：断言关注推送时传给 sender 的 page 为 `pages/notifications/index`；验证配额扣减、fire-and-forget、空 openid/零配额跳过等现有行为不回归。
- 验证命令：`go build ./...` 与 `go test ./internal/wechat/... ./internal/subscribepush/...`。

## Spec Patch

无。delta spec 的验收场景（成功跳转 / page 空值兼容 / 现有行为不变）已在 open 阶段覆盖。

