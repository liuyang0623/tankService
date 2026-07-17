---
comet_change: wechat-subscribe-notification
role: technical-design
canonical_spec: openspec
status: final
archived-with: 2026-07-17-wechat-subscribe-notification
status: final
---

# 微信订阅消息推送 — Design Doc

> Change: wechat-subscribe-notification（Change B）
> 日期: 2026-07-17
> 阶段: design（深度技术设计，细化 open 阶段 design.md）

## 1. 目标与定位

站内通知（notification-center，已上线）的**微信侧增强**：有人关注用户 a 时，除写站内通知外，若 a 有订阅配额，异步向 a 推送一条微信订阅消息。两者并存，推送失败不影响关注与站内通知。

## 2. 架构：follow 双注入（方案 A）

`follow.ToggleFollow` 成功创建关注后，作为"关注事件"源头，分别触发两个独立后果：

```
ToggleFollow (Create 成功)
├── notifier.CreateFollow(...)          // 已有：同步写站内通知，失败记日志
└── subscribePusher.PushFollow(...)     // 新增：异步微信推送，失败记日志
```

- follow service 再注入一个 `subscribePusher` 接口（消费者侧定义，单向依赖，无包循环——与现有 notifier 注入同模式）：
  ```go
  type subscribePusher interface {
      PushFollow(ctx context.Context, targetID, actorID uint)  // 内部自行判配额/查 openid，异步，无返回错误（失败内部记日志）
  }
  func (s *FollowService) SetSubscribePusher(p subscribePusher)
  ```
- `PushFollow` 无 error 返回：推送是 fire-and-forget，成败不回传给关注主流程。follow 侧调用即返回，真正的 goroutine + 超时在 pusher 实现内部。

## 3. 组件

### 3.1 `internal/wechat/`（通用微信能力）
- `type Client struct`：持有 appID/secret、http.Client、access_token 缓存（token + expireAt）、mutex。
- `GetAccessToken(ctx) (string, error)`：缓存未过期直接返回；否则调 `GET /cgi-bin/token?grant_type=client_credential&appid=&secret=`，解析 `{access_token, expires_in, errcode, errmsg}`，缓存 `expires_in - 300s`（留 5 分钟余量）。加锁防并发重复取。
- `SendSubscribeMessage(ctx, openid, tplID string, data map[string]any) error`：取 token → POST `/cgi-bin/message/subscribe/send?access_token=`，body `{touser, template_id, data}`，解析 errcode，非 0 返回错误。
- HTTP 模式照 `internal/auth/service.go` 的 fetchWechatSession（http.Client + JSON 解析 + errcode 判断）。

### 3.2 订阅推送触发器（连接 follow ↔ wechat + 配额）
follow 需要的 `PushFollow` 由一个薄适配器实现（放 `internal/wechat/` 或独立 pusher 文件），职责：
1. goroutine + `context.WithTimeout(5s)` 隔离
2. 查被关注者 openid + `subscribe_follow_quota`
3. quota > 0 → 组装 data → `SendSubscribeMessage` → 成功则 `quota -= 1`（DB update）
4. 任何失败：记日志，返回

### 3.3 配额存储（users 模块）
- User 表加 `SubscribeFollowQuota int` `gorm:"column:subscribe_follow_quota;default:0"`（避开保留字，见 AGENT.md DB 坑）。
- 授权上报：`POST /api/v1/users/subscribe/follow`（JWT）→ `quota += 1`（`UPDATE ... SET subscribe_follow_quota = subscribe_follow_quota + 1`，原子自增）。

### 3.4 前端授权开关（个人中心）
- 个人中心加"关注通知"开关项。
- 点击 → `Taro.requestSubscribeMessage({ tmplIds: [模板ID] })`：
  - 结果 `[模板ID]: 'accept'` → 调 `usersApi.subscribeFollow()` 上报 → toast「已开启，关注提醒将推送一次（微信订阅为一次性，可反复开启累积）」
  - `'reject'` → toast「需授权才能收到关注提醒」
- 模板 ID 放 `src/config/`（与 env.ts 同级），后端从 `WECHAT_SUBSCRIBE_TPL_FOLLOW` 读。

## 4. 数据流

```
[个人中心开关] --requestSubscribeMessage--> 微信授权弹窗
     |同意
     v
POST /users/subscribe/follow --> quota += 1

[用户 b 关注 a]
     v
follow.ToggleFollow → Create 成功
     ├─ notifier.CreateFollow (同步, 站内通知)
     └─ subscribePusher.PushFollow (异步 goroutine)
            ├─ 查 a.openid + a.quota
            ├─ quota>0: wechat.SendSubscribeMessage(a.openid, tpl, data)
            │            └─ 成功: a.quota -= 1
            └─ quota==0 或失败: 记日志, 不影响关注/站内通知
```

## 5. 模板字段（build 阶段需用户提供）

模板 ID `Q2BcepkCFvBshhtriPZlJVWA471xoYkoyJ7xDS4T_BA` 的字段结构（如 `thing1`=关注者昵称、`time2`=关注时间）需从微信公众平台模板详情获取。`data` 按 `{ "thing1": {"value": nickname}, "time2": {"value": timeStr} }` 组装。build 阶段 task 4.1 向用户索取准确字段名。

## 6. 错误处理与边界

| 情况 | 处理 |
|------|------|
| access_token 获取失败 | SendSubscribeMessage 返回错误 → pusher 记日志，不推 |
| 推送 goroutine 超时 | 5s context 超时 → 记日志 |
| quota == 0 | 不推送，站内通知照常 |
| openid 为空（异常） | 跳过推送，记日志 |
| 推送成功但 quota 扣减失败 | 记日志（可能多扣一次配额，可接受） |
| 关注自己 | ToggleFollow 已拦截，不触发 |

## 7. 测试策略

- `internal/wechat`：mock http.Client，测 token 缓存命中/过期刷新、SendSubscribeMessage 成功/errcode 失败。
- 配额上报：handler 层 mock，测 quota+1、未登录 401。
- follow 触发：mock subscribePusher，验证关注成功调用 PushFollow、取关不调用。（真实推送靠真机联调。）

## 8. 部署与回滚

- AutoMigrate 加 `subscribe_follow_quota` 列，重启后端建列。
- 回滚：移除 wechat 包 + 授权接口 + follow 的 pusher 注入；字段残留无害。

## 9. 已知约束

- **一次性订阅**：配额耗尽收不到，文案说清。
- **access_token 单实例内存缓存**：多实例部署需换共享缓存（Redis 等），本期单实例不做。

