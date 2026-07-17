# Implementation Tasks — fix-interactions-bugs

## 1. 点赞/收藏 toggle 500 修复

- [x] 1.1 LikePost 删除分支改 `Unscoped().Delete`（物理删除）
- [x] 1.2 FavoritePost 删除分支改 `Unscoped().Delete`
- [x] 1.3 复现验证：同一帖子反复点赞/取消 3+ 次不再 500

## 2. 评论 DTO

- [x] 2.1 定义 `CommentResponse` DTO（小写字段 + author{id,name,avatar}）
- [x] 2.2 GetPostComments 返回 DTO（组装作者，含 replies 作者）
- [x] 2.3 CreateComment 返回 DTO（重新查作者信息）
- [x] 2.4 验证评论列表/创建返回含 author 与小写字段

## 3. 可选鉴权中间件

- [x] 3.1 新增 `OptionalJWTMiddleware`（有效 token 注入 userID，否则放行）
- [x] 3.2 挂到公开路由（详情/列表/评论列表/用户帖子）
- [x] 3.3 验证：带 token 详情返回真实 isLiked/isFavorited；无 token 正常返回

## 4. 验证

- [x] 4.1 go build 通过；相关单元测试通过
- [x] 4.2 重启服务，curl 复现三个场景全部正常
