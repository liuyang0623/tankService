# Implementation Tasks — message（后端私信能力）

## 1. 依赖与中间件提取

- [x] 1.1 `go get github.com/gorilla/websocket`
- [x] 1.2 `pkg/middleware/jwt.go`：抽出 `ParseToken(tokenString, secret) (*Claims, error)` 供 WS 复用（从 `JWTMiddleware` 中提取逻辑）

## 2. 数据模型

- [x] 2.1 `internal/message/model.go`：`Conversation`（UserAID/UserBID/LastMessage/LastTime/UserAUnread/UserBUnread，`(user_a_id,user_b_id)` 唯一索引）、`Message`（ConversationID/SenderID/Type/Content/Read，ConversationID 索引）
- [x] 2.2 约定：创建会话时保证较小 ID 为 UserA、较大为 UserB

## 3. WebSocket Hub

- [x] 3.1 `internal/message/hub.go`：Hub 结构（`map[uint]*Client` + sync.RWMutex）、Client 结构（UserID/Conn/Send chan）
- [x] 3.2 `NewHub()` + `Run()`（启动 Hub goroutine，读取 client 注册/注销 channel）
- [x] 3.3 `SendToUser(userID uint, data []byte)`：向用户所有在线连接推送
- [x] 3.4 `ServeWS(conn, userID)`：升级后处理——注册 client、启动读/写 goroutine，读循环接收并忽略消息（仅 pong handler），写循环从 Send chan 消费写入 WS
- [x] 3.5 心跳：用 `conn.SetPingHandler` 或 goroutine 定期 ping（每 54s），pong 等待 10s，超时关闭

## 4. Service 层

- [x] 4.1 `internal/message/service.go`：`MessageService` 依赖 `*gorm.DB`
- [x] 4.2 `CreateMessage(ctx, senderID, toUserID uint, msgType, content string) (*Message, error)`：查找/创建 Conversation（规范化排序），创建 Message，更新 Conversation 的 LastMessage/LastTime/对方未读数，返回 Message
- [x] 4.3 `GetConversations(ctx, userID uint, page, limit int)`：分页查当前用户参与的会话，填充对方用户信息（id/nickname/avatar）、自己的未读数
- [x] 4.4 `GetMessages(ctx, conversationID, userID uint, page, limit int) (*PaginatedMessages, error)`：验证用户是该会话参与者，分页查消息（按 `created_at DESC`）
- [x] 4.5 `MarkRead(ctx, conversationID, userID uint) error`：验证参与关系，将该会话中当前用户为接收方的消息标记已读，重置未读数

## 5. Handler 层

- [x] 5.1 `internal/message/handler.go`：`MessageHandler` 结构 + `NewMessageHandler(service, hub)`
- [x] 5.2 `SendMessage(c)`：解析 body（toUserId/type/content），调 service.CreateMessage，成功后通过 hub.SendToUser 推送
- [x] 5.3 `ListConversations(c)`：分页返回会话列表
- [x] 5.4 `GetMessages(c)`：分页返回历史消息
- [x] 5.5 `MarkRead(c)`：标记已读
- [x] 5.6 `HandleWS(c)`：从 `?token=` 解析 JWT → 调 hub.ServeWS → gorilla/websocket 升级
- [x] 5.7 复用 `parsePagination` 和 `getUserID` 模式（来自 follow handler）

## 6. 注册到 main.go

- [x] 6.1 `cmd/server/main.go`：`setupRouter` 中注册路由组：
  - `authorized.GET("/conversations", msgHandler.ListConversations)`
  - `authorized.GET("/conversations/:id/messages", msgHandler.GetMessages)`
  - `authorized.POST("/messages", msgHandler.SendMessage)`
  - `authorized.POST("/conversations/:id/read", msgHandler.MarkRead)`
  - 根路由（无 JWT 前缀）：`r.GET("/ws", msgHandler.HandleWS)`（使用 OptionalJWTMiddleware 解析 token）
- [x] 6.2 `main()` 中：创建 `hub := message.NewHub()` + `go hub.Run()` + `AutoMigrate` 注册模型
- [x] 6.3 将 Hub 实例传给 `NewMessageHandler`

## 7. 测试

- [x] 7.1 `internal/message/handler_test.go`：发送消息、会话列表、历史消息、标记已读、WS 连接拒绝
- [x] 7.2 `bun test`（或 `go test ./...`）通过
- [x] 7.3 tsc（实际是 go vet + go build）通过
