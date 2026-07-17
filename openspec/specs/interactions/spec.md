# interactions Specification

## Purpose
TBD - created by archiving change go-service. Update Purpose after archive.
## Requirements
### Requirement: 点赞/取消点赞
系统 SHALL 允许已认证用户对文章点赞；再次请求 SHALL 取消点赞（toggle 语义）；likeCount 计数实时更新。

#### Scenario: 点赞文章
- **WHEN** POST `/api/v1/posts/:id/like`，用户未点赞该文章
- **THEN** 创建点赞记录，likeCount +1，返回 `{liked: true}`

#### Scenario: 取消点赞
- **WHEN** POST `/api/v1/posts/:id/like`，用户已点赞
- **THEN** 删除点赞记录，likeCount -1，返回 `{liked: false}`

### Requirement: 收藏/取消收藏
系统 SHALL 允许已认证用户收藏文章；再次请求 SHALL 取消收藏（toggle 语义）。

#### Scenario: 收藏文章
- **WHEN** POST `/api/v1/posts/:id/favorite`，用户未收藏
- **THEN** 创建收藏记录，返回 `{favorited: true}`

#### Scenario: 取消收藏
- **WHEN** POST `/api/v1/posts/:id/favorite`，用户已收藏
- **THEN** 删除收藏记录，返回 `{favorited: false}`

### Requirement: 获取用户收藏列表
系统 SHALL 提供当前用户的收藏文章分页列表。

#### Scenario: 获取收藏列表
- **WHEN** GET `/api/v1/users/me/favorites?page=1&limit=10`
- **THEN** 返回收藏的文章摘要列表（含 favoritedAt）

### Requirement: 创建评论
系统 SHALL 允许已认证用户对文章发表评论；支持回复其他评论（parentId）；commentCount 实时更新。

#### Scenario: 发表顶级评论
- **WHEN** POST `/api/v1/comments`，body 含 postId 和 content
- **THEN** 创建评论，commentCount +1，返回评论信息（含作者）

#### Scenario: 回复评论
- **WHEN** POST `/api/v1/comments`，body 含 postId、content、parentId
- **THEN** 创建子评论，commentCount +1

### Requirement: 获取文章评论列表
系统 SHALL 返回文章的顶级评论列表（分页），每条评论包含其回复列表。

#### Scenario: 获取评论
- **WHEN** GET `/api/v1/posts/:id/comments?page=1&limit=10`
- **THEN** 返回顶级评论（含 replies），按创建时间倒序

### Requirement: 删除评论
系统 SHALL 允许评论作者删除自己的评论（级联删除回复），commentCount 相应减少。

#### Scenario: 作者删除评论
- **WHEN** DELETE `/api/v1/comments/:id`，当前用户为评论作者
- **THEN** 删除评论及其回复，commentCount 递减

#### Scenario: 非作者删除
- **WHEN** DELETE `/api/v1/comments/:id`，当前用户非作者
- **THEN** 返回 403

