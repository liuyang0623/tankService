# Implementation Tasks — follow（go-service 后端关注能力）

## 1. Follow 模型与模块骨架

- [x] 1.1 新增 `internal/follow/model.go`：Follow 结构（follower_id/following_id 复合唯一索引）+ TableName
- [x] 1.2 `internal/follow/service.go`：FollowService 骨架 + NewFollowService

## 2. 关注/取关 toggle

- [x] 2.1 `ToggleFollow(ctx, followerID, targetID)`：校验不能关注自己、目标存在，硬删/Create toggle
- [x] 2.2 `handler.go`：ToggleFollow handler（从 JWT 取 followerID，路径取 targetID）

## 3. 粉丝/关注列表

- [x] 3.1 `ListFollowers`/`ListFollowing`：join users 分页查询 + FollowUserItem DTO
- [x] 3.2 isFollowing 批量计算（一次 IN 查询当前用户对本页用户的关注关系，避免 N+1）
- [x] 3.3 handler：ListFollowers/ListFollowing（OptionalJWT 取 currentUserID）

## 4. 用户详情带关注计数

- [x] 4.1 users 模块 GetUser 返回 UserDetailResponse（加 followerCount/followingCount/isFollowing）
- [x] 4.2 计数查询：users 轻量直查 follows 表（避免循环依赖）；isFollowing 按当前登录用户算
- [x] 4.3 路由 GET /users/:id 从 public 移到 OptionalJWT 组

## 5. 路由注册与 migration

- [x] 5.1 `cmd/server/main.go`：注册 follow 路由（POST follow 受保护，followers/following/GET :id 用 OptionalJWT）
- [x] 5.2 AutoMigrate 加 `&follow.Follow{}`

## 6. 验证

- [x] 6.1 单元测试：ToggleFollow（关注/取关/不能关注自己）、列表分页、计数
- [x] 6.2 `go build ./...` 编译通过
- [x] 6.3 `go test ./...` 通过
- [x] 6.4 curl 冒烟：关注/取关、粉丝列表、关注列表、用户详情带计数（本地起服务）
