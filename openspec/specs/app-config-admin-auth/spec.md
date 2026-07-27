# app-config-admin-auth Specification

## Purpose
TBD - created by archiving change add-appconfig-admin-auth. Update Purpose after archive.
## Requirements
### Requirement: 审核模式切换接口需通过管理员白名单鉴权

系统 SHALL 提供 `PUT /api/v1/app-config` 接口用于切换审核模式开关(`auditMode`),该接口 MUST 先通过 JWT 登录校验,再通过管理员白名单校验后才允许执行更新。管理员白名单 MUST 由环境变量 `ADMIN_OPENIDS`(逗号分隔的 openid 列表)配置,系统 MUST 以当前登录用户 `userID` 反查其 `openid` 并与白名单比对。

#### Scenario: 白名单管理员切换审核模式成功

- **WHEN** 已登录用户的 `openid` 存在于 `ADMIN_OPENIDS` 白名单,且携带合法 JWT 调用 `PUT /api/v1/app-config` 传入 `auditMode`
- **THEN** 系统更新单行配置的 `audit_mode` 列并返回成功,后续 `GET /api/v1/app-config` 返回更新后的值

#### Scenario: 非白名单用户被拒绝

- **WHEN** 已登录用户携带合法 JWT,但其 `openid` 不在 `ADMIN_OPENIDS` 白名单,调用 `PUT /api/v1/app-config`
- **THEN** 系统返回 403 Forbidden,且不修改任何配置

#### Scenario: 未登录请求被拦截

- **WHEN** 请求未携带合法 JWT 调用 `PUT /api/v1/app-config`
- **THEN** 系统在 JWT 中间件层返回 401,不进入管理员白名单校验

