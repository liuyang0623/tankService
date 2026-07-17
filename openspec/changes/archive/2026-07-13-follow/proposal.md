## Why

「摆烂随笔」小程序要增加用户关注功能——他人主页可关注/取关、查看关注列表与粉丝列表、个人中心展示关注数/被关注数。这些前端功能都依赖 go-service 提供关注关系的存储与查询。本 change 是「用户社交」四拆分的第 1 个（后端关注能力），前端 user-follow change 消费本契约。

## What Changes

- **关注关系表**：新增 `follows`（follower_id → following_id，唯一索引防重复，不能关注自己）
- **关注/取关**：`POST /users/:id/follow` toggle（受保护，已关注则取关）
- **粉丝列表**：`GET /users/:id/followers` 分页返回关注 ta 的用户
- **关注列表**：`GET /users/:id/following` 分页返回 ta 关注的用户
- **用户详情扩展**：`GET /users/:id` 响应增加 `followerCount`/`followingCount`/`isFollowing`（当前登录用户是否已关注 ta；可选鉴权，未登录时 isFollowing=false）
- **列表项 DTO**：粉丝/关注列表返回精简用户信息（id/nickname/avatar/bio/isFollowing）
- AutoMigrate 注册 `Follow` 模型

## Capabilities

### New Capabilities

- `user-follow`: 用户关注能力——关注/取关、粉丝列表、关注列表、用户详情带关注计数与关注态

### Modified Capabilities

<!-- 无。GET /users/:id 增加字段属向后兼容扩展，不改既有响应结构的已有字段。 -->

## Impact

- **新增模块**：`internal/follow/`（model.go / service.go / handler.go）
- **修改**：`internal/users/`（GetUser 响应加关注计数/关注态，或由 follow 模块提供聚合）、`cmd/server/main.go`（注册路由 + AutoMigrate Follow）
- **复用**：现有 `PaginationMeta` 分页风格、`JWTMiddleware`/`OptionalJWTMiddleware`（列表 isFollowing 个性化）、User 模型
- **数据库**：新增 `follows` 表（migration 通过 AutoMigrate）
- **前端契约**：供 user-follow change 消费；GET /users/:id 加字段对现有前端 getUser 兼容
