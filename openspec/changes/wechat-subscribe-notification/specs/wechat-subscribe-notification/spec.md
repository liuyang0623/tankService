# wechat-subscribe-notification Specification

## ADDED Requirements

### Requirement: 微信 access_token 获取与缓存
系统 SHALL 提供获取微信 access_token 的能力，并缓存至其过期前复用，避免每次请求都调用微信接口（微信有频率限制）。

#### Scenario: 首次获取 access_token
- **WHEN** 缓存为空或已过期时请求 access_token
- **THEN** 调用微信 `cgi-bin/token` 获取并缓存，返回 token

#### Scenario: 缓存有效期内复用
- **WHEN** 缓存中的 access_token 未过期
- **THEN** 直接返回缓存值，不调用微信接口

### Requirement: 订阅授权配额上报
系统 SHALL 提供接口让登录用户上报订阅授权，每次授权成功使其可推送配额加一。

#### Scenario: 授权上报累加配额
- **WHEN** 登录用户调用授权上报接口
- **THEN** 该用户的 `subscribe_follow_quota` 加一

#### Scenario: 未登录上报被拒
- **WHEN** 未携带有效 JWT 调用上报接口
- **THEN** 返回 401

### Requirement: 关注时按配额推送订阅消息
当有用户关注某用户时，若被关注者有剩余订阅配额，系统 SHALL 通过微信订阅消息接口向其推送一条关注通知，并将配额减一。推送为异步执行，失败或配额不足 MUST NOT 影响关注操作与站内通知写入。

#### Scenario: 有配额时推送
- **WHEN** 用户 b 关注用户 a，且 a 的订阅配额大于 0
- **THEN** 系统向 a 推送微信订阅消息，a 的配额减一

#### Scenario: 无配额时不推送
- **WHEN** 用户 b 关注用户 a，但 a 的订阅配额为 0
- **THEN** 系统不推送订阅消息，但站内通知仍正常写入

#### Scenario: 推送失败不影响关注
- **WHEN** 关注已成功，但订阅消息推送发生错误或超时
- **THEN** 关注操作与站内通知不受影响，仅记录错误日志

#### Scenario: 取消关注不推送
- **WHEN** 用户对已关注的对象执行取消关注
- **THEN** 不推送订阅消息

### Requirement: 个人中心订阅授权开关
前端 SHALL 在个人中心提供"关注通知"授权开关。用户点击时 SHALL 调用微信 `requestSubscribeMessage`；用户同意后 SHALL 上报后端累加配额，并提示订阅为一次性（授权一次推送一条）。

#### Scenario: 用户开启授权
- **WHEN** 用户在个人中心点击"关注通知"开关
- **THEN** 弹出微信订阅授权；用户同意后上报后端，配额加一，并提示一次性订阅说明

#### Scenario: 用户拒绝授权
- **WHEN** 用户在订阅授权弹窗中拒绝
- **THEN** 不上报后端，提示需授权才能收到关注提醒
