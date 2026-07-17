## Why

关注能力已上线，用户可以互相关注，但还不能私下交流。私信是社交闭环的关键一环。用户明确选择 WebSocket 实时方案（而非轮询），因为后续计划接入小程序订阅消息提醒服务，实时推送架构能平滑衔接。本 change 交付后端私信能力：会话/消息持久化 + WebSocket 实时收发。

## What Changes

- **新增 message 模块**（`internal/message/`）：会话（Conversation）与消息（Message）模型、服务、HTTP handler、WebSocket Hub
- **数据模型**：
  - `Conversation`：双方用户会话（user_a_id/user_b_id 规范化排序 + 复合唯一索引），维护最后消息、更新时间
  - `Message`：会话内消息（conversation_id、sender_id、type=text|image、content、已读标记）
- **REST 接口**（JWT 保护）：
  - `GET /conversations`：当前用户会话列表（含对方用户信息、最后一条消息、未读数），分页
  - `GET /conversations/:id/messages`：某会话历史消息，分页
  - `POST /messages`：发送消息（body: toUserId, type, content）→ 落库 + 通过 Hub 实时推送给在线的对方
  - `POST /conversations/:id/read`：标记会话已读
- **WebSocket 接口**：
  - `GET /ws?token=<jwt>`：升级为 WebSocket 连接，query 带 token 握手鉴权（小程序 wx.connectSocket 无法自定义 header）
  - 内存 Hub 管理 userID→连接映射；收到新消息时向在线接收方推送 JSON 帧
  - 心跳保活（ping/pong），断线清理
- **依赖**：引入 `github.com/gorilla/websocket`
- **AutoMigrate**：新增 `&message.Conversation{}`、`&message.Message{}`

## Capabilities

### New Capabilities
- `messaging`: 后端私信能力——会话与消息持久化、REST 收发与历史查询、WebSocket 实时推送、未读计数

### Modified Capabilities
<!-- 无既有 spec 需求变更 -->

## Impact

- **新增模块**：`internal/message/`（model.go / service.go / handler.go / hub.go / handler_test.go）
- **修改**：`cmd/server/main.go`（注册 REST + WS 路由、AutoMigrate、启动 Hub goroutine）、`go.mod`（+gorilla/websocket）
- **鉴权**：WS 用 query token；复用 `pkg/middleware` 的 token 解析逻辑（抽出可复用的 ParseToken）
- **实时范围**：单实例内存 Hub，多实例水平扩展需 Redis pub/sub（本 change 非目标，YAGNI）
- **前端**：本 change 不含前端；前端私信 UI 在后续 change④（user-messaging）
- **图片消息**：content 存图片 URL（前端先走既有 upload 接口拿 URL 再发 type=image 消息），后端不处理图片上传本身
