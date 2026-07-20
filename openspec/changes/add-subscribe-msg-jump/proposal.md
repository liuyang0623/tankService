## Why

用户在微信收到「关注」订阅消息后点击，当前会落到小程序首页，而不是相关的系统通知页面。微信订阅消息接口 `subscribe/send` 支持 `page` 参数指定点击后打开的小程序页面，但现有后端下发时未携带该字段，导致跳转能力缺失。本次补齐后端跳转能力，让用户点击关注订阅消息后直达系统通知列表页。

## What Changes

- `wechat.Client.SendSubscribeMessage` 的请求 body 增加 `page` 字段，支持透传微信订阅消息点击后的跳转页；`page` 为空时不写入该字段，保持与现状一致的兼容行为。
- `subscribepush.Pusher` 推送关注订阅消息时，将 `page` 设为前端真实存在的系统通知页 `pages/notifications/index`，用户点击后直达系统通知列表。
- 跳转页路径 `pages/notifications/index` 作为后端与小程序端的对接契约固化在设计文档中。

### 非目标（明确不做）

- 不新增单条通知详情接口。前端系统通知页 `pages/notifications/index` 只依赖现有 `GET /notifications` 列表接口自给自足，深链进入也只需拉列表即可渲染。
- 不修改通知数据模型、不新增通知类型。
- 不修改订阅配额扣减、授权或 openid 逻辑。
- 不改动前端小程序代码（前端页面已存在）。

## Capabilities

### New Capabilities

无新增 capability。

### Modified Capabilities

- `subscribe-push`: 关注订阅消息下发时携带跳转页 `page`，指向系统通知页。
