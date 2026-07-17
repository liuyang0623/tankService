# post-category Specification

## Purpose
TBD - created by archiving change post-category-search. Update Purpose after archive.
## Requirements
### Requirement: 文章分类字段

系统 SHALL 为文章提供固定分类字段 `category`，合法值为 story/daily/tech/food/travel，可空（兼容旧数据）。

#### Scenario: 发布带分类的文章

- **WHEN** 用户发布文章并指定合法分类
- **THEN** 系统 SHALL 保存该分类到文章

#### Scenario: 非法分类值

- **WHEN** 发布文章指定非法分类值
- **THEN** 系统 SHALL 返回 400 错误

#### Scenario: 不带分类发布

- **WHEN** 用户发布文章不指定分类
- **THEN** 系统 SHALL 允许 category 为空

### Requirement: 分类列表接口

系统 SHALL 通过 `GET /categories` 返回固定分类列表，每项含 value 和 label。

#### Scenario: 获取分类列表

- **WHEN** 客户端请求 GET /categories
- **THEN** 系统 SHALL 返回固定的 5 个分类（value + 中文 label），无需鉴权

