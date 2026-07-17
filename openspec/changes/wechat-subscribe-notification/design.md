## Context

站内通知（notification-center）已上线，关注时写站内通知的触发点在 `internal/follow/service.go` 的 `ToggleFollow`（Create 成功后经注入的 notifier 触发）。本 change 在此基础上增加微信订阅消息推送。微信 HTTP 调用模式已有范例：`internal/auth/service.go` 的 `fetchWechatSession`（http.Client + errcode 解析）。微信订阅消息是一次性订阅，授权一次只能推一条。

## Goals / Non-Goals

**Goals:**
- 后端能获取并缓存微信 access_token、发送订阅消息。
- 用户可在个人中心主动授权，后端按配额推送关注通知。
- 推送为站内通知的增强，失败不影响关注与站内通知。

**Non-Goals:**
- 点赞/评论的订阅推送（仅关注）。
- 公众号模板消息（仅小程序订阅消息）。
- 多实例共享 access_token 缓存（本期单实例内存缓存）。

## Decisions

### 1. 新增 `internal/wechat/` 包，复用现有微信调用模式
封装两件事：`GetAccessToken(ctx)`（`cgi-bin/token`，内存缓存 + 过期刷新，微信限流要求缓存 ~7200s，不能每次现取）、`SendSubscribeMessage(ctx, openid, tplID, data)`（`cgi-bin/message/subscribe/send`）。HTTP 调用照 `internal/auth/service.go` 的 http.Client + errcode 解析模式。
**备选**：在 auth 包里加——弃，职责不同，wechat 通用能力独立成包更清晰。

### 2. 配额模型：用户表加 `subscribe_follow_quota`
一次性订阅无法覆盖所有粉丝，用配额攒次数。授权上报 `quota += 1`，推送成功 `quota -= 1`。openid 已落库（`internal/users/model.go` 有 `Openid`），复用。
**备选**：每次进小程序自动补授权——弃，弹窗扰民；用户主动开关体验更可控。

### 3. 关注触发：异步推送 + 错误隔离
`ToggleFollow` 成功创建关注后，除已有的同步写站内通知外，**异步**（goroutine + 超时 context）触发订阅推送：查被关注者 openid + `quota > 0` → 发送 → 成功 `quota -= 1`。失败只记日志。
**为何异步**（与站内通知不同）：微信 API 是慢 I/O（网络往返 + 可能超时），不能拖慢关注主流程；站内通知是本地 DB 写入，同步即可。
依赖注入：follow service 再注入一个 subscribe pusher 接口（单向，避免包循环），与现有 notifier 注入模式一致。

### 4. 授权上报接口
`POST /api/v1/users/subscribe/follow`（JWT）：前端 `requestSubscribeMessage` 用户同意后调用，`quota += 1`。放 users 模块（配额是用户属性）。

### 5. 前端授权开关放个人中心
个人中心页加"关注通知"开关项，点击 → `Taro.requestSubscribeMessage({tmplIds:[模板ID]})` → 同意则上报后端 + toast 说明"一次性订阅，关注提醒推送一次"。模板 ID 放 `src/config/`。

## Risks / Trade-offs

- [一次性订阅配额耗尽收不到] → 开关文案说清"授权一次推一条"；站内通知兜底，不依赖微信推送覆盖全部。
- [access_token 每次现取触发微信限流] → 内存缓存 ~7200s + 提前刷新；单实例够用，多实例部署需换共享缓存（本期注明不做）。
- [异步推送 goroutine 泄漏/阻塞] → 带超时 context，失败快速返回只记日志。
- [模板字段填错被微信拒发] → build 阶段从公众平台取准确字段结构（thing/time/name），按模板要求填。

## Migration Plan

- 用户表加 `subscribe_follow_quota int default 0`：随 AutoMigrate 建列，需重启后端。纯新增无数据迁移。
- 回滚：删 wechat 包 + 授权接口 + follow 里的推送调用；字段残留无害。

## Open Questions

- **模板字段结构**（thing1/time2/name1 等具体字段名与顺序）——build 阶段需用户从微信公众平台模板详情提供，否则推送内容无法正确填充。
