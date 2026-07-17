# message Specification

## Purpose
提供后端私信能力，支撑用户之间的私密交流。包含会话管理、消息收发历史、未读计数，以及 WebSocket 实时推送。前端在后续 change（user-messaging）接入。

## Requirements

### Requirement: 私信发送

系统 SHALL 允许已登录用户通过 `POST /messages` 向其他用户发送消息，自动创建/查找会话，消息落库，并实时推送给在线的接收方。

#### Scenario: 发送文字消息（已有会话）

- **WHEN** 已登录用户向有过会话的用户发送文字消息
- **THEN** 系统 SHALL 将消息存入该会话，更新会话最后消息/时间及接收方未读数，并实时推送

#### Scenario: 发送文字消息（首次）

- **WHEN** 已登录用户向从未会话过的用户发送文字消息
- **THEN** 系统 SHALL 创建新会话，再执行消息发送流程

#### Scenario: 发送图片消息

- **WHEN** 已登录用户发送 type=image 的消息体（content 为图片 URL）
- **THEN** 系统 SHALL 同文字消息流程存储并推送

#### Scenario: 不能给自己发消息

- **WHEN** 发送方 toUserId 等于自己的 id
- **THEN** 系统 SHALL 返回 400 Bad Request

### Requirement: WebSocket 实时推送

系统 SHALL 提供一个 WebSocket 端点 `GET /ws?token=<jwt>`，通过 query 携带 JWT 完成鉴权，在用户在线时实时推送新消息。

#### Scenario: 建立连接

- **WHEN** 客户端携带有效 JWT token 请求 WebSocket 升级
- **THEN** 系统 SHALL 鉴权通过，建立连接并注册到 Hub

#### Scenario: 鉴权失败

- **WHEN** 客户端携带无效或过期 token 请求升级
- **THEN** 系统 SHALL 拒绝连接

#### Scenario: 推送新消息

- **WHEN** 用户 A 发送消息给用户 B 且 B 在线
- **THEN** 系统 SHALL 通过 B 的 WebSocket 连接推送包含消息完整信息的 JSON 帧

#### Scenario: 断线清理

- **WHEN** 用户 WebSocket 连接断开（心跳超时或主动关闭）
- **THEN** 系统 SHALL 从 Hub 移除该连接

#### Scenario: 多端在线

- **WHEN** 同一用户在不同设备建立多个 WebSocket 连接
- **THEN** 系统 SHALL 向该用户的全部连接推送消息

### Requirement: 历史消息列表

系统 SHALL 通过 `GET /conversations` 返回当前登录用户的会话列表，分页，按最后消息时间倒序，每项含对方用户信息、最后一条消息、最后时间、未读数。

#### Scenario: 查看会话列表

- **WHEN** 已登录用户请求 GET /conversations
- **THEN** 系统 SHALL 返回分页会话列表，含对方 id/昵称/头像、最后消息、未读数

#### Scenario: 无会话

- **WHEN** 从未发过消息的用户请求会话列表
- **THEN** 系统 SHALL 返回空列表

#### Scenario: 查询与某用户的会话

- **WHEN** 已登录用户请求 `GET /conversations?withUser=<userId>`
- **THEN** 系统 SHALL 返回与该用户的会话 id（`{conversationId}`），无会话时返回 0

#### Scenario: 图片消息摘要

- **WHEN** 会话最后一条消息为图片
- **THEN** 系统 SHALL 在会话列表 lastMessage 字段返回 `[图片]` 占位符而非图片 URL

### Requirement: 历史消息详情

系统 SHALL 通过 `GET /conversations/:id/messages` 返回某会话的消息历史，分页，按时间倒序。

#### Scenario: 查看历史消息

- **WHEN** 已登录用户请求某会话的历史消息
- **THEN** 系统 SHALL 返回分页消息列表（id/senderId/type/content/createdAt）

#### Scenario: 查看非自己的会话

- **WHEN** 用户请求不属于自己的会话消息
- **THEN** 系统 SHALL 返回 403 Forbidden

### Requirement: 标记已读

系统 SHALL 通过 `POST /conversations/:id/read` 允许用户将某会话中所有未读消息标记为已读。

#### Scenario: 标记已读

- **WHEN** 用户请求标记某会话为已读
- **THEN** 系统 SHALL 将该会话中当前用户作为接收方的消息全部设为已读，重置未读数
