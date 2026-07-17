# users Specification

## Purpose
TBD - created by archiving change go-service. Update Purpose after archive.
## Requirements
### Requirement: 获取当前用户个人资料
系统 SHALL 允许已认证用户获取自己的个人资料（不含密码字段）。

#### Scenario: 获取个人资料
- **WHEN** GET `/api/v1/users/profile`，携带有效 JWT
- **THEN** 返回当前用户信息（id, name, avatar, phone, email, openid, createdAt）

### Requirement: 更新当前用户个人资料
系统 SHALL 允许已认证用户更新自己的 name、avatar、phone 字段。

#### Scenario: 更新成功
- **WHEN** PATCH `/api/v1/users/profile`，body 含需更新字段
- **THEN** 返回更新后的用户信息

### Requirement: 查询指定用户信息
系统 SHALL 允许查询任意用户的公开信息（id, name, avatar）。

#### Scenario: 用户存在
- **WHEN** GET `/api/v1/users/:id`，用户 ID 有效
- **THEN** 返回该用户公开信息

#### Scenario: 用户不存在
- **WHEN** GET `/api/v1/users/:id`，ID 不存在
- **THEN** 返回 404

