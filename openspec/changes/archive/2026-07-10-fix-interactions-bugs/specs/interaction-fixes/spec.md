## ADDED Requirements

### Requirement: 点赞/收藏可无限次 toggle

系统 SHALL 允许用户对同一帖子无限次点赞/取消点赞、收藏/取消收藏，不因历史软删除记录与唯一索引冲突而报错。

#### Scenario: 反复点赞取消不报错

- **WHEN** 用户对同一帖子连续多次点赞、取消、再点赞
- **THEN** 系统 SHALL 每次都正确返回当前状态（`{liked}`），不返回 500

#### Scenario: 反复收藏取消不报错

- **WHEN** 用户对同一帖子连续多次收藏、取消、再收藏
- **THEN** 系统 SHALL 每次都正确返回当前状态（`{favorited}`），不返回 500

### Requirement: 评论返回带作者的 DTO

评论列表与创建接口 SHALL 返回小写 json 字段并包含作者信息（id、昵称、头像）的 DTO，而非 GORM 原始模型。

#### Scenario: 评论列表含作者

- **WHEN** 客户端请求某帖子的评论列表
- **THEN** 每条评论 SHALL 包含 `id`、`content`、`author{id,name,avatar}`、`parentId`、`replies`、`createdAt` 等小写字段

#### Scenario: 创建评论返回含作者

- **WHEN** 用户成功发表评论
- **THEN** 返回的评论对象 SHALL 包含作者信息与小写字段，可直接用于前端回显

### Requirement: 公开路由可选鉴权

系统 SHALL 为公开的详情/列表路由提供可选鉴权：请求携带有效 token 时解析并注入当前用户，未携带或无效时放行为匿名。

#### Scenario: 登录用户获取详情互动态

- **WHEN** 已登录用户携带有效 token 请求帖子详情
- **THEN** 系统 SHALL 返回该用户对该帖的真实 `isLiked`/`isFavorited`

#### Scenario: 匿名用户正常浏览

- **WHEN** 未登录用户请求帖子详情（无 token）
- **THEN** 系统 SHALL 正常返回详情，`isLiked`/`isFavorited` 为 false，不报错
