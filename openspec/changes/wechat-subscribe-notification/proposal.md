## Why

站内通知（notification-center）只在用户打开小程序时可见，离开后无感知。有人关注用户时，希望能通过微信订阅消息推送到微信，做到"离开小程序也能收到"，提升关注互动的触达率。

## What Changes

- 新增**微信订阅消息推送能力**：后端封装微信 access_token 获取（带缓存）与订阅消息发送。
- **授权入口**：个人中心新增"关注通知"开关，用户主动点击触发 `requestSubscribeMessage`；同意后上报后端累加配额。
- **配额模型**：微信订阅消息是一次性订阅（授权一次只能推一条），用户表记录剩余可推次数（`subscribe_follow_quota`），授权一次 +1，推送一条 -1。
- **关注时推送**：有人关注用户且被关注者有配额时，异步调用微信订阅消息接口推送；配额耗尽则不推（站内通知仍照常写入）。
- 推送为**站内通知的增强**，不替代——两者并存。

## Capabilities

### New Capabilities
- `wechat-subscribe-notification`: 微信订阅消息推送能力——access_token 缓存获取、订阅消息发送、订阅授权配额管理、关注事件触发推送、前端授权开关。

### Modified Capabilities
<!-- 关注行为（user-follow）对外契约不变，仅在已有的"关注成功后写站内通知"之外附带一次订阅推送，属实现细节，不改 spec 级要求。notification-center 也不改，订阅推送是独立增强层。 -->

## Impact

- **后端（tankService）**：新增 `internal/wechat/`（access_token 缓存 + SendSubscribeMessage）；用户表加 `subscribe_follow_quota` 字段 + migration；新增授权上报接口 `POST /users/subscribe/follow`；`internal/follow/service.go` 关注成功后异步触发订阅推送（复用已注入的通知触发点）。
- **前端（小程序）**：个人中心加"关注通知"开关 → `Taro.requestSubscribeMessage` → 上报后端；模板 ID 入 `src/config/`。
- **配置**：模板 ID `Q2Bce...T_BA` 配到后端环境变量 `WECHAT_SUBSCRIBE_TPL_FOLLOW` + 前端 config；复用已有 `WECHAT_APPID`/`WECHAT_SECRET`。
- **依赖**：无新增第三方依赖；复用现有微信 HTTP 调用模式（`internal/auth/service.go`）、JWT 鉴权、请求层。
- **约束**：一次性订阅——配额耗尽收不到；access_token 需缓存 ~7200s（单实例内存缓存，多实例需共享缓存，本期单实例）。
