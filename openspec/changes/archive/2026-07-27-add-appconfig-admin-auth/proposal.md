## Why

`GET /api/v1/app-config` 已能下发审核模式开关,但运营切换该开关目前没有可用且受控的写接口。审核模式直接影响全端互动板块的显隐,必须限定只有管理员本人能改,不能让任意登录用户调用。因此需要新增一个带管理员白名单鉴权的审核模式切换接口。

## What Changes

- 新增 `PUT /api/v1/app-config` 接口,用于切换 `auditMode`(审核模式开关)。
- 接口挂在 `authorized` 分组下:先经 `JWTMiddleware` 校验登录态,再经新增的管理员白名单中间件校验权限。
- 新增管理员白名单中间件:从 `gin.Context` 取 JWT 注入的 `userID`,按 `userID` 反查 `users.openid`,与配置白名单比对,不在白名单则返回 403。
- `config.Config` 新增 `AdminOpenids` 字段,从环境变量 `ADMIN_OPENIDS`(逗号分隔)加载并解析为集合。
- `AppConfigService` 新增更新单行 `audit_mode` 的方法;`AppConfigHandler` 新增对应 PUT handler。
- 补充三个单测覆盖:白名单命中放行、非白名单拒绝(403)、审核模式更新落库。

## Capabilities

### New Capabilities

- `app-config-admin-auth`: 管理员白名单鉴权下的应用配置写能力,覆盖审核模式切换接口的鉴权规则与更新行为。

### Modified Capabilities

<!-- 现有 openspec/specs/ 下无 app-config 相关 capability,GET 接口未建 spec。本次不修改既有 spec 的验收场景,故留空。 -->

## Impact

- 代码:`pkg/config/config.go`(新增配置字段与解析)、`internal/appconfig/service.go`、`internal/appconfig/handler.go`(新增更新方法与 PUT handler)、新增管理员鉴权中间件、`cmd/server/main.go`(装配 PUT 路由 + 中间件)。
- 配置:新增环境变量 `ADMIN_OPENIDS`(逗号分隔的管理员 openid 列表)。
- API:新增 `PUT /api/v1/app-config`;既有 `GET /api/v1/app-config` 行为不变。
- 依赖:管理员中间件需按 `userID` 读取 `users` 表的 `openid` 列。
