# diary Specification

## ADDED Requirements

### Requirement: 日记创建

系统 SHALL 允许已登录用户通过 `POST /diaries` 创建私密日记，含标题、富文本内容、封面、心情、天气、多图。

#### Scenario: 创建日记

- **WHEN** 已登录用户提交日记内容
- **THEN** 系统 SHALL 保存日记（归属该用户）并返回详情

#### Scenario: 未登录创建

- **WHEN** 未登录请求创建日记
- **THEN** 系统 SHALL 返回 401

### Requirement: 日记列表（时间线）

系统 SHALL 通过 `GET /diaries` 返回当前用户的日记列表，分页，按创建时间倒序，仅本人可见。

#### Scenario: 查看我的日记

- **WHEN** 已登录用户请求日记列表
- **THEN** 系统 SHALL 返回该用户的日记分页列表（含心情/天气/摘要/封面）

#### Scenario: 只见自己的日记

- **WHEN** 用户请求日记列表
- **THEN** 系统 SHALL 仅返回归属该用户的日记

### Requirement: 日记详情

系统 SHALL 通过 `GET /diaries/:id` 返回日记详情，仅归属本人可访问。

#### Scenario: 查看自己的日记详情

- **WHEN** 用户请求自己日记的详情
- **THEN** 系统 SHALL 返回完整内容（标题/富文本/图片/心情/天气）

#### Scenario: 访问他人日记

- **WHEN** 用户请求不属于自己的日记
- **THEN** 系统 SHALL 返回 404（不泄露存在性）

### Requirement: 日记更新与删除

系统 SHALL 通过 `PATCH /diaries/:id` 更新、`DELETE /diaries/:id` 删除日记，均仅归属本人可操作。

#### Scenario: 更新自己的日记

- **WHEN** 用户更新自己的日记
- **THEN** 系统 SHALL 保存变更（含重建图片）

#### Scenario: 删除自己的日记

- **WHEN** 用户删除自己的日记
- **THEN** 系统 SHALL 删除日记及关联图片

#### Scenario: 操作他人日记

- **WHEN** 用户更新或删除不属于自己的日记
- **THEN** 系统 SHALL 返回 404
