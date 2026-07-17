# 验证报告 — wechat-subscribe-notification

> Change: wechat-subscribe-notification（微信订阅消息推送）
> 验证级别: full（tasks=19，delta specs=1，changed files=21，均达 full 阈值）
> 验证日期: 2026-07-17

## 1. 验证范围

站内通知的微信侧增强：有人关注用户时，除写站内通知外，若被关注者有订阅配额，异步推送一条微信订阅消息。覆盖后端能力包、配额模型、授权上报接口、关注触发推送，以及前端授权开关与真机端到端。

## 2. 自动化验证

| 项 | 命令 | 结果 |
|----|------|------|
| 编译 | `go build ./...` | ✅ 通过 |
| 全量测试 | `go test ./...` | ✅ 全绿 |

关键测试覆盖：

- `internal/wechat`：access_token 缓存命中/过期刷新、SendSubscribeMessage 成功/errcode 失败路径（mock HTTP）。
- `internal/subscribepush`：有配额推送、无配额跳过、openid 为空跳过、成功后扣减配额、PushFollow 异步不阻塞。
- `internal/users`：授权上报累加配额（`IncrSubscribeFollowQuota`）、`SubscribeFollow` handler 成功与未登录 401。
- `internal/follow`：ToggleFollow 成功触发 PushFollow、取关不触发（mock pusher）。

## 3. 手动/真机验证

- 端到端（真机，用户执行）：个人中心开授权 → 微信弹订阅 → 同意 → 换号关注该用户 → 微信收到订阅消息 → 配额耗尽不再推。**用户确认通过。**
- 模板字段：`thing1`=关注者昵称（≤20 字符截断）、`time2`=关注时间，真机推送成功，字段与公众平台模板一致。

## 4. 需求符合性

| 需求 | 状态 |
|------|------|
| 微信 access_token 缓存获取 | ✅ `internal/wechat/client.go` |
| 订阅消息发送（errcode 解析） | ✅ `SendSubscribeMessage` |
| 配额模型（授权 +1 / 推送 -1） | ✅ `subscribe_follow_quota` + Incr/Decr |
| 授权上报接口 `POST /users/subscribe/follow`（JWT） | ✅ |
| 关注成功异步推送、失败不影响主流程 | ✅ fire-and-forget + 超时 ctx |
| 站内通知与订阅推送并存 | ✅ 双注入（notifier + pusher） |
| 文档同步（README / .env.example） | ✅ 后端 README 三段 + env 模板 |

## 5. 已知约束与残留

- **一次性订阅**：配额耗尽收不到，前端文案已说明。
- **access_token 单实例内存缓存**：多实例部署需换共享缓存（Redis），本期单实例不做。
- **AutoMigrate 只增不删**：`subscribe_follow_quota` 列建列需重启后端，回滚时字段残留无害。

## 6. 结论

自动化验证（build + test）全绿，真机端到端由用户确认通过，需求全部满足。**验证通过（pass）。**
