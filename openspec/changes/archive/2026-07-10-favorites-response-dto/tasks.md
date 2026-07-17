# Implementation Tasks — favorites-response-dto

## 1. posts 公开转换
- [x] 1.1 posts 包新增公开 `ToPostResponse(post Post) PostResponse`

## 2. 收藏 DTO
- [x] 2.1 `FavoriteItem.Post` 改 `posts.PostResponse`
- [x] 2.2 GetUserFavorites Preload Author/Images/Topics + 逐条转换

## 3. 验证
- [x] 3.1 go build + 相关测试通过
- [x] 3.2 重启后 curl 验证收藏返回小写 DTO（需 token，或前端真机验证）
