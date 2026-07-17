# user-follow Specification

## Purpose
TBD - created by archiving change follow. Update Purpose after archive.
## Requirements
### Requirement: 关注与取关

系统 SHALL 允许已登录用户关注或取关其他用户，通过 `POST /users/:id/follow` 切换关注状态；不允许关注自己；重复关注不产生重复记录。

#### Scenario: 关注用户

- **WHEN** 已登录用户对未关注的目标用户发起 follow
- **THEN** 系统 SHALL 建立关注关系并返回 following=true

#### Scenario: 取关用户

- **WHEN** 已登录用户对已关注的目标用户再次发起 follow
- **THEN** 系统 SHALL 删除关注关系并返回 following=false

#### Scenario: 不能关注自己

- **WHEN** 用户对自己的 id 发起 follow
- **THEN** 系统 SHALL 拒绝并返回错误

### Requirement: 粉丝列表

系统 SHALL 通过 `GET /users/:id/followers` 分页返回关注该用户的用户列表，每项含 id、昵称、头像、简介，以及当前登录用户对该项用户的关注态（isFollowing）。

#### Scenario: 查看粉丝列表

- **WHEN** 请求某用户的粉丝列表
- **THEN** 系统 SHALL 分页返回关注 ta 的用户，含精简信息与 isFollowing

### Requirement: 关注列表

系统 SHALL 通过 `GET /users/:id/following` 分页返回该用户关注的用户列表，字段同粉丝列表项。

#### Scenario: 查看关注列表

- **WHEN** 请求某用户的关注列表
- **THEN** 系统 SHALL 分页返回 ta 关注的用户，含精简信息与 isFollowing

### Requirement: 用户详情带关注计数

系统 SHALL 在 `GET /users/:id` 响应中返回该用户的粉丝数 followerCount、关注数 followingCount，以及当前登录用户是否已关注该用户 isFollowing（可选鉴权，未登录时 isFollowing=false）。

#### Scenario: 登录用户查看他人详情

- **WHEN** 已登录用户请求他人 `GET /users/:id`
- **THEN** 系统 SHALL 返回 followerCount、followingCount 与 isFollowing（反映当前用户是否已关注 ta）

#### Scenario: 未登录查看他人详情

- **WHEN** 未登录请求 `GET /users/:id`
- **THEN** 系统 SHALL 返回 followerCount、followingCount，isFollowing 为 false

