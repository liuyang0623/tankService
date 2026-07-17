# posts Specification

## Purpose
TBD - created by archiving change go-service. Update Purpose after archive.
## Requirements
### Requirement: 创建文章
系统 SHALL 允许已认证用户创建文章，支持草稿（DRAFT）和直接发布（PUBLISHED）两种状态，文章可关联图片 URL 列表和话题标签。

#### Scenario: 创建草稿
- **WHEN** POST `/api/v1/posts`，body 含 title、content，status 为 DRAFT
- **THEN** 创建文章，返回完整文章信息（含 author、images、topics）

#### Scenario: 直接发布
- **WHEN** POST `/api/v1/posts`，status 为 PUBLISHED
- **THEN** 创建文章并设置 publishedAt 为当前时间

### Requirement: 获取已发布文章列表

系统 SHALL 提供已发布文章的分页列表，默认按发布时间倒序，每条返回摘要信息（前100字符）、封面图（最多3张）和分类 category。列表 SHALL 支持按关键词、分类、排序方式、关注关系筛选。

#### Scenario: 分页列表

- **WHEN** GET `/api/v1/posts?page=1&limit=10`
- **THEN** 返回 `{data: [...], meta: {total, page, limit, totalPages}}`，每项含 category

#### Scenario: 关键词搜索

- **WHEN** GET `/api/v1/posts?keyword=旅行`
- **THEN** 系统 SHALL 返回 title 包含"旅行"的已发布文章

#### Scenario: 分类过滤

- **WHEN** GET `/api/v1/posts?category=story`
- **THEN** 系统 SHALL 返回该分类的已发布文章

#### Scenario: 推荐排序

- **WHEN** GET `/api/v1/posts?sort=likes`
- **THEN** 系统 SHALL 按点赞数倒序返回文章

#### Scenario: 关注流

- **WHEN** 已登录用户 GET `/api/v1/posts?following=true`
- **THEN** 系统 SHALL 仅返回该用户关注的作者的已发布文章

#### Scenario: 关注流未登录

- **WHEN** 未登录请求 `/api/v1/posts?following=true`
- **THEN** 系统 SHALL 返回空列表

#### Scenario: 组合筛选

- **WHEN** GET `/api/v1/posts?category=tech&keyword=go`
- **THEN** 系统 SHALL 返回同时满足分类和关键词的文章

### Requirement: 获取文章详情
系统 SHALL 返回文章完整内容，并增加 viewCount；若当前用户已登录，附带 isLiked/isFavorited 状态；草稿只有作者可查看。

#### Scenario: 查看已发布文章
- **WHEN** GET `/api/v1/posts/:id`
- **THEN** 返回完整文章信息，viewCount +1

#### Scenario: 非作者查看草稿
- **WHEN** GET `/api/v1/posts/:id`，文章为 DRAFT，当前用户非作者
- **THEN** 返回 403

### Requirement: 更新文章
系统 SHALL 允许作者更新自己的文章，支持更新标题、内容、状态、图片、话题；状态从 DRAFT→PUBLISHED 时自动设置 publishedAt。

#### Scenario: 作者更新文章
- **WHEN** PATCH `/api/v1/posts/:id`，当前用户为作者
- **THEN** 返回更新后的完整文章信息

#### Scenario: 非作者更新
- **WHEN** PATCH `/api/v1/posts/:id`，当前用户非作者
- **THEN** 返回 403

### Requirement: 删除文章
系统 SHALL 允许作者删除自己的文章（级联删除图片、话题关联、点赞、收藏、评论）。

#### Scenario: 作者删除
- **WHEN** DELETE `/api/v1/posts/:id`，当前用户为作者
- **THEN** 删除文章，返回成功消息

### Requirement: 发布文章
系统 SHALL 允许作者将草稿文章发布。

#### Scenario: 发布草稿
- **WHEN** POST `/api/v1/posts/:id/publish`，文章为 DRAFT
- **THEN** 状态改为 PUBLISHED，设置 publishedAt

### Requirement: 用户文章列表
系统 SHALL 提供当前用户的草稿箱（仅 DRAFT）和全部文章（含 DRAFT）两个分页接口。

#### Scenario: 草稿箱
- **WHEN** GET `/api/v1/posts/drafts/my`，携带有效 JWT
- **THEN** 返回当前用户的草稿列表

#### Scenario: 全部文章
- **WHEN** GET `/api/v1/posts/my`，携带有效 JWT
- **THEN** 返回当前用户所有文章（含草稿）

