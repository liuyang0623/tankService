# diary-notebook Specification

## Purpose
TBD - created by archiving change diary-notebook. Update Purpose after archive.
## Requirements
### Requirement: 日记本管理

系统 SHALL 允许已登录用户创建、查看、更新、删除私密日记本，每个日记本含名称、颜色、可选封面。

#### Scenario: 创建日记本

- **WHEN** 已登录用户提交日记本名称和颜色
- **THEN** 系统 SHALL 创建日记本（归属该用户）并返回详情

#### Scenario: 查看我的日记本

- **WHEN** 已登录用户请求日记本列表
- **THEN** 系统 SHALL 返回该用户全部日记本（含每本日记数）

#### Scenario: 自动创建默认本

- **WHEN** 用户首次请求日记本列表且无任何日记本
- **THEN** 系统 SHALL 自动创建一个"默认"日记本并返回

#### Scenario: 更新/删除日记本

- **WHEN** 用户更新或删除自己的日记本
- **THEN** 系统 SHALL 保存变更或删除该日记本

#### Scenario: 操作他人日记本

- **WHEN** 用户更新或删除不属于自己的日记本
- **THEN** 系统 SHALL 返回 404

#### Scenario: 删除日记本保留日记

- **WHEN** 用户删除含日记的日记本
- **THEN** 系统 SHALL 删除日记本但保留其中日记（notebook_id 置 0）

### Requirement: 日记归属日记本

系统 SHALL 让日记归属某个日记本，并支持按日记本过滤日记列表。

#### Scenario: 创建日记指定日记本

- **WHEN** 用户创建日记并指定 notebookId
- **THEN** 系统 SHALL 将日记归属该日记本

#### Scenario: 按日记本过滤

- **WHEN** 用户请求日记列表并带 notebookId 参数
- **THEN** 系统 SHALL 仅返回该日记本内的日记

