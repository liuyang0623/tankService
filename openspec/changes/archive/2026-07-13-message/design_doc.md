# Design Doc — 后端私信能力（message）

## 1. 架构概述

```
┌─────────────────┐   REST /api/v1/*     ┌──────────────────────┐
│  微信小程序       │◄────────────────────►│  Gin Server (:3000)   │
│  (Taro/前端)      │                      │                      │
│                  │   WebSocket ws://*    │  ┌────────────────┐  │
│                  │◄────────────────────►│  │  Hub (内存)      │  │
│                  │                      │  │  userID→WS conn  │  │
│                  │                      │  └────────────────┘  │
└─────────────────┘                      └──────────┬───────────┘
                                                    │ GORM
                                                    ▼
                                           ┌──────────────────┐
                                           │   MySQL           │
                                           │  conversations    │
                                           │  messages         │
                                           └──────────────────┘
```

### 核心设计决策

**WebSocket 方案**：单实例内存 Hub，不引入 Redis。当前一个实例够用，且用户选了「gorilla/websocket + 内存 Hub」。后续水平扩展时换 Redis pub/sub 即可（Hub 替换为接口实现）。

**消息模型**：Conversation 会话表（双方 user 规范化排序 + 复合唯一索引）+ Message 消息表。会话列表查询快，未读数好维护。

**消息类型**：text / image 两种。emoji 当 text 存（Unicode），无需特殊处理。

**鉴权**：WebSocket 握手走 `?token=<jwt>` query 参数。小程序 `wx.connectSocket` 无法自定义 header，只能走 query。复用 `pkg/middleware` 的 token 解析逻辑（抽出 `ParseToken()` 函数）。

## 2. 数据模型

### Conversation 会话

```go
type Conversation struct {
    gorm.Model
    UserAID      uint      `gorm:"not null;index"`
    UserBID      uint      `gorm:"not null;index"`
    LastMessage  string    `gorm:"type:text"`
    LastTime     time.Time
    UserAUnread  int       `gorm:"default:0"`  // UserA 的未读消息数
    UserBUnread  int       `gorm:"default:0"`  // UserB 的未读消息数
}
```

- `UserAID` / `UserBID` 按较小 ID 为 A、较大 ID 为 B，保证 `(UserAID, UserBID)` 唯一
- `gorm:uniqueIndex:idx_conversation_pair` 防重复会话
- 双方发送第一条消息时自动创建

### Message 消息

```go
type Message struct {
    gorm.Model
    ConversationID uint      `gorm:"not null;index"`
    SenderID       uint      `gorm:"not null"`
    Type           string    `gorm:"type:varchar(20);not null;default:'text'"`
    Content        string    `gorm:"type:text;not null"`
    Read           bool      `gorm:"default:false"`
}
```

- `type`: `text` / `image`
- `content`: 文字内容或图片 URL
- `Read`: 是否已读（标记已读时批量更新）
- `ConversationID` 索引 + `sender_id` 可联合索引（已有 `gorm.Model` 的 `ID` + `CreatedAt` 倒序查分页）

## 3. REST API 设计

所有 REST 接口需要 JWT 鉴权（`authorized` 路由组）。

### GET /conversations — 会话列表

- 分页（page/limit），按 `last_time DESC` 排序
- 返回：对方用户信息（id/nickname/avatar）、最后一条消息、最后时间、未读数
- 自己的未读数：如果当前用户是 UserA 则取 `UserAUnread`，否则取 `UserBUnread`

### GET /conversations/:id/messages — 历史消息

- 分页（page/limit），按 `created_at DESC` 排序（倒序查，前端反转）
- 返回消息列表：id/senderId/type/content/createdAt

### POST /messages — 发送消息

- Body: `{"toUserId": uint, "type": "text|image", "content": "string"}`
- 流程：
  1. 查找或创建 Conversation（保证较小 ID 为 A）
  2. 创建 Message 记录
  3. 更新 Conversation 的 `LastMessage` / `LastTime` / 对方未读数
  4. 通过 Hub 推送实时消息给接收方（如果在线）
- 返回：message 对象

### POST /conversations/:id/read — 标记已读

- 将当前用户在该会话中作为接收方的未读消息全部标记为 `Read: true`
- 重置 Conversation 中当前用户的未读数

## 4. WebSocket Hub 设计

### Hub 结构

```go
type Hub struct {
    clients map[uint]*Client  // userID → Client
    mu      sync.RWMutex
}

type Client struct {
    UserID uint
    Conn   *websocket.Conn
    Send   chan []byte
}
```

### 连接生命周期

1. **握手**: 客户端连接 `/ws?token=<jwt>` → 解析 token 拿到 userID → 升级为 WS → 注册到 Hub
2. **读循环**: 读消息（当前仅用于接收 ping/pong），收到 `CloseMessage` 时清理
3. **写循环**: 从 `Send` chan 取数据写入 WS 连接
4. **心跳**: server 每 54s 发 ping，等待 pong（超时 10s 断开）
5. **断线**: 连接关闭时从 Hub 移除

### 消息推送格式

```json
{
  "type": "new_message",
  "data": {
    "id": 1,
    "conversationId": 1,
    "senderId": 2,
    "type": "text",
    "content": "你好",
    "createdAt": "2026-07-13T12:00:00+08:00"
  }
}
```

### 消息流

```
用户 A POST /messages → 服务端落库 → Hub.SendToUser(UserB, messageJSON)
                                    → 如果 UserB 在线，写入其 WebSocket
                                    → 如果离线，消息存库下次打开 WS 可拉历史
```

## 5. Spec 变更

### 新增 Capabilities

- `message`: 后端私信能力——会话与消息持久化、REST 收发与历史查询、WebSocket 实时推送、未读计数

### 变更模块

| 模块 | 类型 | 说明 |
|------|------|------|
| `internal/message/` | 新增 | model.go / service.go / handler.go / hub.go / handler_test.go |
| `cmd/server/main.go` | 修改 | 注册路由、AutoMigrate、启动 Hub |
| `go.mod` | 修改 | +gorilla/websocket |

## 6. 边界与限制

- **单实例**：内存 Hub 只在当前进程内通信。多实例部署需改为 Redis pub/sub，本 change 不做。
- **已读回执**：仅标记会话所有未读为已读，不支持单条消息回执。
- **删除消息**：本 change 不实现消息撤回或删除。
- **图片上传**：复用既有的 `/upload/image` 接口上传拿到 URL，再发 type=image 消息。后端不处理图片上传本身。
- **敏感词过滤**：本 change 不做。
