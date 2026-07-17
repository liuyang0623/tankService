# posts Specification

## MODIFIED Requirements

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
