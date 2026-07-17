## ADDED Requirements

### Requirement: 微信小程序登录
系统 SHALL 接受微信小程序登录请求，通过 `code` 换取 `openid`，查找或创建用户，并签发 JWT Token。

#### Scenario: 新用户首次登录
- **WHEN** POST `/api/v1/auth/wechat/login`，body 含有效 `code`
- **THEN** 调用微信 jscode2session 接口，创建新用户，返回 JWT token 和用户基本信息

#### Scenario: 老用户再次登录
- **WHEN** POST `/api/v1/auth/wechat/login`，body 含有效 `code`，openid 已存在
- **THEN** 更新用户 session_key，返回 JWT token 和用户基本信息

#### Scenario: 微信 code 无效
- **WHEN** POST `/api/v1/auth/wechat/login`，`code` 无效或已过期
- **THEN** 返回 401，错误信息说明登录失败

### Requirement: JWT 认证中间件
系统 SHALL 提供 JWT 中间件，保护需要认证的路由；未携带有效 token 的请求 SHALL 被拒绝。

#### Scenario: 携带有效 Token
- **WHEN** 请求头包含 `Authorization: Bearer <valid_token>`
- **THEN** 中间件通过，将 userID 注入 Gin context，请求继续处理

#### Scenario: Token 缺失或无效
- **WHEN** 请求未携带 token 或 token 已过期
- **THEN** 返回 401 Unauthorized
