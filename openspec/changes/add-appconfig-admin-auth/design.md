## Context

`GET /api/v1/app-config`(免鉴权)已下发 `auditMode`。运营需要一个受控写接口切换该开关,但只能由管理员本人操作。现有基础设施:`pkg/middleware/jwt.go` 的 `JWTMiddleware` 校验后 `c.Set("userID", claims.UserID)`;`internal/users/model.go` 的用户表含 `Openid` 唯一字段;`pkg/config/config.go` 用 `os.Getenv` 集中加载配置;路由在 `cmd/server/main.go` 的 `authorized` 分组统一挂 JWT。

## Goals / Non-Goals

**Goals:**
- 新增 `PUT /api/v1/app-config` 切换 `auditMode`,受管理员白名单保护。
- 白名单基于 openid,通过环境变量配置,新增管理员无需改代码。
- 复用现有 JWT 中间件,鉴权分两层:先登录、后白名单。

**Non-Goals:**
- 不引入角色/权限表或 RBAC 体系(白名单足够当前单一管理员场景)。
- 不改动既有 `GET /api/v1/app-config` 行为。
- 不做管理员操作审计日志(超出本次范围)。

## Decisions

- **白名单存储方式**:用环境变量 `ADMIN_OPENIDS`(逗号分隔)而非数据库表。理由:管理员数量极少且稳定,环境变量零 schema 变更、部署即生效;RBAC 表在当前需求下过度设计。`config.Load` 解析为 `[]string`(或 set),空值时白名单为空,任何用户都无法通过,安全默认关闭。
- **鉴权在 openid 而非 userID**:userID 是自增主键,跨环境不稳定;openid 是微信侧稳定唯一标识,配置可跨环境复用。中间件按 `c.GetUint("userID")` 反查 `users.openid` 再比对。
- **两层中间件顺序**:`authorized` 分组已统一 `JWTMiddleware`;在其后再挂一个 `AdminOnly` 中间件,仅作用于该 PUT 路由。未登录由 JWT 层返回 401,登录但非白名单由 AdminOnly 返回 403,职责清晰。
- **中间件依赖注入**:`AdminOnly` 需要 db(反查 openid)和白名单集合。构造为 `AdminOnlyMiddleware(db, adminOpenids)` 返回 `gin.HandlerFunc`,在 main.go 装配时注入,便于单测。
- **Service 更新方法**:新增 `UpdateAuditMode(ctx, bool)`,对单行(`singletonID`)做 upsert 语义的更新,复用既有单行常量与表结构。

## Risks / Trade-offs

- [环境变量未配置导致无人可操作] → 属安全默认(fail-closed);部署文档说明必须配置 `ADMIN_OPENIDS`。
- [每次 PUT 都查一次 users 表] → 写接口调用频率极低,单次主键查询开销可忽略,不做缓存。
- [openid 拼写错误导致管理员被拒] → 通过单测覆盖命中/拒绝两条路径,并在错误响应中返回明确 403 便于排查。

## Migration Plan

- 部署前在环境变量中配置 `ADMIN_OPENIDS=oFJcJ48bX2RwRs0StgvYUldHNGMA`。
- 无数据库 schema 变更,无需 migration。
- 回滚:移除 PUT 路由与中间件即可,GET 接口与配置行不受影响。
