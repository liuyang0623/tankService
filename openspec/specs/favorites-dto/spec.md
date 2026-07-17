# favorites-dto Specification

## Purpose
TBD - created by archiving change favorites-response-dto. Update Purpose after archive.
## Requirements
### Requirement: 收藏返回规范 DTO

收藏列表接口 SHALL 返回 `{post, favoritedAt}`，其中 `post` 为 PostResponse DTO（小写 json 字段，含 author、images、topics），与帖子列表一致。

#### Scenario: 收藏列表 post 为规范 DTO

- **WHEN** 客户端请求我的收藏列表
- **THEN** 每项的 `post` SHALL 为小写字段的 PostResponse，含 `author{id,name,avatar}`

#### Scenario: 空收藏

- **WHEN** 用户无收藏
- **THEN** 系统 SHALL 返回空 data 与正确 meta，不报错

