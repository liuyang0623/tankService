## Why

收藏列表接口 `GET /users/me/favorites` 返回的 `FavoriteItem.post` 是 GORM 原始模型 `posts.Post`（PascalCase 字段、无 author 信息），与其它列表接口（PostResponse 规范 DTO）不一致，前端无法直接消费。小程序个人中心「我的收藏」需要规范结构，为保持数据一致性，统一改为返回 PostResponse DTO。

## What Changes

- 在 posts 包暴露公开转换函数 `ToPostResponse(post Post) PostResponse`（复用现有 `toPostResponse` 逻辑）
- `GetUserFavorites` 查帖子时 Preload Author/Images/Topics，并将 `FavoriteItem.Post` 从 `posts.Post` 改为 `posts.PostResponse`
- 收藏返回结构对齐帖子列表（小写字段 + author）

## Capabilities

### New Capabilities

- `favorites-dto`: 收藏列表返回规范 PostResponse DTO（小写字段 + author），与帖子列表一致

### Modified Capabilities

无

## Impact

- `internal/posts/service.go`：新增公开 `ToPostResponse`
- `internal/interactions/service.go`：`FavoriteItem.Post` 改 DTO，Preload 关联
- 无 schema 变更；前端 `user-profile-center` 依赖本修复消费规范收藏结构
