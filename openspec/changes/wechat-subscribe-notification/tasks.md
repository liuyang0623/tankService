# wechat-subscribe-notification 实施任务

## 1. 后端 wechat 能力包

- [ ] 1.1 新建 `internal/wechat/`：`GetAccessToken(ctx)`（调 cgi-bin/token，内存缓存 + 过期刷新，缓存 ~7200s）
- [ ] 1.2 `SendSubscribeMessage(ctx, openid, tplID, data)`（调 cgi-bin/message/subscribe/send，errcode 解析），HTTP 模式参照 internal/auth/service.go
- [ ] 1.3 单测：access_token 缓存命中/过期刷新；发送成功/失败路径（mock HTTP）

## 2. 订阅配额与授权上报

- [ ] 2.1 用户表加 `subscribe_follow_quota int default 0`，加入 AutoMigrate
- [ ] 2.2 users 模块加 `POST /api/v1/users/subscribe/follow`（JWT）：quota += 1
- [ ] 2.3 单测：授权上报累加配额、未登录 401

## 3. 关注触发订阅推送

- [ ] 3.1 follow service 注入 subscribe pusher 接口（单向依赖，参照现有 notifier 注入）
- [ ] 3.2 ToggleFollow 成功后异步（goroutine + 超时 ctx）：查被关注者 openid + quota>0 → SendSubscribeMessage → 成功 quota-=1；失败记日志
- [ ] 3.3 装配：main.go 构造 wechat service 注入 follow；模板 ID 从 env `WECHAT_SUBSCRIBE_TPL_FOLLOW` 读取
- [ ] 3.4 单测：有配额推送、无配额不推、推送失败不影响关注返回

## 4. 模板字段对接（需用户提供）

- [ ] 4.1 从微信公众平台取模板 `Q2Bce...T_BA` 的字段结构（thing/time/name 等）
- [ ] 4.2 按字段结构组装 SendSubscribeMessage 的 data（关注者昵称 + 时间等）

## 5. 后端验证

- [ ] 5.1 `go build ./...` 通过
- [ ] 5.2 `go test ./...` 全绿

## 6. 前端授权开关

- [ ] 6.1 模板 ID 入 `src/config/`；`src/services/api/users.ts` 加 `subscribeFollow()` 上报方法
- [ ] 6.2 个人中心加"关注通知"开关：点击 `Taro.requestSubscribeMessage({tmplIds})` → 同意则上报 + toast 一次性订阅说明；拒绝则提示
- [ ] 6.3 前端验证：tsc + vitest + build:weapp

## 7. 端到端与收尾

- [ ] 7.1 端到端（真机）：个人中心开授权 → 微信弹订阅 → 同意 → 换号关注该用户 → 微信收到订阅消息 → 配额耗尽不再推
- [ ] 7.2 按 AGENT.md 文档同步铁律更新前后端 README（订阅消息能力）
