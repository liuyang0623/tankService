## Context

go-service 增加关注能力。关注关系与现有 Like/Favorite 同构（`internal/interactions`）——都是唯一索引关系表 + toggle + 计数，参照其实现。路由注册在 `cmd/server/main.go` 的 `setupRouter`，AutoMigrate 在 `main()`。User 模型在 `internal/users`。分页复用 `PaginationMeta`。

## Goals / Non-Goals

**Goals:** 关注/取关 toggle、粉丝列表、关注列表、用户详情带关注计数与关注态。

**Non-Goals:** 不做前端；不做私信；不做关注动态流/推荐/互关标识；不做软删除（关注取关直接硬删，与 fix-interactions-bugs 后 Like/Favorite 一致）。

## Decisions

**D1. Follow 模型（新模块 internal/follow）**
```go
type Follow struct {
    gorm.Model
    FollowerID  uint `gorm:"not null;index:idx_follow_pair,unique"` // 关注发起者
    FollowingID uint `gorm:"not null;index:idx_follow_pair,unique"` // 被关注者
}
func (Follow) TableName() string { return "follows" }
```
- 复合唯一索引 `idx_follow_pair(follower_id, following_id)` 防重复关注
- 硬删除取关（避免软删除 + 唯一索引冲突，同 Like/Favorite 修复经验）

**D2. FollowService（参照 InteractionService）**
- `ToggleFollow(ctx, followerID, targetID) (following bool, err error)`：
  - targetID == followerID → 返回错误「不能关注自己」
  - 校验目标用户存在
  - 查已有关系：存在则硬删（取关 false），不存在则 Create（关注 true）
- `ListFollowers(ctx, userID, currentUserID, page, limit) (PaginatedUsers, error)`：join users 取关注 userID 的人
- `ListFollowing(ctx, userID, currentUserID, page, limit) (PaginatedUsers, error)`：join users 取 userID 关注的人
- `CountFollowers/CountFollowing(userID) int64`、`IsFollowing(currentUserID, targetID) bool` — 供用户详情聚合

**D3. 列表项 DTO**
```go
type FollowUserItem struct {
    ID          uint   `json:"id"`
    Nickname    string `json:"nickname"`
    Avatar      string `json:"avatar"`
    Bio         string `json:"bio"`
    IsFollowing bool   `json:"isFollowing"` // 当前登录用户是否关注该项用户
}
type PaginatedUsers struct {
    Data []FollowUserItem `json:"data"`
    Meta PaginationMeta   `json:"meta"`
}
```
- isFollowing 批量计算：查当前用户对本页所有用户的关注关系，避免 N+1

**D4. 用户详情扩展（GET /users/:id）**
- 现有 `users.GetUser` 返回 User。扩展为返回带计数的 DTO：
```go
type UserDetailResponse struct {
    // ...原 User 字段（id/nickname/avatar/bio/gender 等）
    FollowerCount  int64 `json:"followerCount"`
    FollowingCount int64 `json:"followingCount"`
    IsFollowing    bool  `json:"isFollowing"`
}
```
- 路由改用 `OptionalJWTMiddleware`（现在是纯 public），登录时算 isFollowing，未登录 isFollowing=false
- follow 计数由 FollowService 提供，users handler 注入 FollowService 或 users 直接查 follows 表（倾向后者：users 模块加轻量 count 查询，避免循环依赖）

**D5. 路由注册（main.go setupRouter）**
- `authorized.POST("/users/:id/follow", followHandler.ToggleFollow)`（受保护）
- 列表用 OptionalJWT（需要 isFollowing 个性化）：
  - `optionalAuth.GET("/users/:id/followers", followHandler.ListFollowers)`
  - `optionalAuth.GET("/users/:id/following", followHandler.ListFollowing)`
- `GET /users/:id` 从 public 移到 optionalAuth 组
- AutoMigrate 加 `&follow.Follow{}`

## Risks / Trade-offs

| 风险 | 缓解 |
|------|------|
| 循环依赖 follow↔users | users 详情直接轻量查 follows 表算计数，不 import follow 包；或 follow 提供函数由 main 组装 |
| isFollowing N+1 | 列表批量查当前用户对本页用户的关注关系，一次 IN 查询 |
| GET /users/:id 从 public 改 optionalAuth | OptionalJWT 未登录也放行，兼容；isFollowing 未登录为 false |
| 关注自己/关注不存在用户 | ToggleFollow 前置校验 |
| 取关软删除唯一索引冲突 | 硬删除（同 Like/Favorite 修复） |

## Migration Plan

AutoMigrate 自动建 `follows` 表。无破坏性变更，GET /users/:id 加字段向后兼容。

## Open Questions

- users 详情算计数：users 直接查 follows 表 vs 注入 FollowService —— 倾向 users 轻量直查避免循环依赖（build 阶段定）
- 关注列表排序：按关注时间倒序（最近关注在前）
