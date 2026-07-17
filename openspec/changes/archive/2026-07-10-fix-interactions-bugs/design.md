## Context

小程序联调发现三个后端交互缺陷，根因均已确证。go-service 基于 Go+Gin+GORM+MySQL。

## Goals / Non-Goals

**Goals:** 修复点赞/收藏 toggle 500、评论 DTO、公开路由可选鉴权。
**Non-Goals:** 不改 URL 契约、不加评论点赞接口（前端本地预留）、不改数据库 schema（唯一索引保留）。

## Decisions

**D1. 点赞/收藏改硬删除**
- 根因：`gorm.Model` 软删除 + `(post_id,user_id)` 唯一索引。软删除记录占用索引，第三次 `Create` 冲突 500。
- 方案：toggle 的删除分支用 `s.db.Unscoped().Delete(&like)`（物理删除），彻底释放唯一索引。
- 备选：查询时 `Unscoped()` 找回软删除记录并复用——较复杂，硬删除更直接（点赞记录无保留软删除的业务价值）。

**D2. 评论 DTO**
- 定义 `CommentResponse{ id, content, authorId, author{id,name,avatar}, parentId, replies[], createdAt }`，service 层 Preload/组装作者后返回。
- GetPostComments 与 CreateComment 均返回该 DTO（CreateComment 需重新查作者）。

**D3. 可选鉴权中间件**
- `pkg/middleware/OptionalJWTMiddleware`：有 `Authorization: Bearer` 且 token 有效则 `c.Set("userID", uid)`；无 token 或无效则不设、放行（不返回 401）。
- 挂到公开路由 `GET /posts`、`GET /posts/:id`、`GET /posts/:id/comments`、`GET /users/:id/posts`。

## Risks / Trade-offs

- [硬删除丢失点赞历史] → 点赞/收藏无历史审计需求，可接受
- [可选鉴权对无效 token 的处理] → 无效 token 一律当匿名放行，不报错，避免影响公开浏览
- [评论 DTO 递归组装作者性能] → 评论量级小，Preload 足够；后续量大再优化

## Migration Plan

无 schema 迁移。重启服务生效。回滚：revert 分支。

## Open Questions

- 评论 replies 的作者是否也需递归组装（本次一并处理顶层与一层回复的作者）
