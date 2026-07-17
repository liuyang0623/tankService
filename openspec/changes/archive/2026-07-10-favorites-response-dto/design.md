## Context
收藏返回原始 GORM 模型，与其它列表 DTO 不一致。posts 有私有 `toPostResponse`，PostResponse 类型公开。

## Goals / Non-Goals
**Goals:** 收藏 post 返回 PostResponse DTO。
**Non-Goals:** 不改 schema、不改 URL。

## Decisions
- **D1**：posts 包新增公开 `ToPostResponse(post Post) PostResponse`，包装现有私有逻辑（或将私有改为调用公开）。
- **D2**：`GetUserFavorites` 查帖子 `Preload("Author").Preload("Images").Preload("Topics")`，`FavoriteItem.Post` 改 `posts.PostResponse`，逐条转换。

## Risks / Trade-offs
- [Preload 增加查询] → 收藏量小，可接受。

## Migration Plan
无 schema 迁移，重启生效。
