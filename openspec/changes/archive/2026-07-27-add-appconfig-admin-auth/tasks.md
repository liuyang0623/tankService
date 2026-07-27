## 1. 配置加载

- [x] 1.1 `pkg/config/config.go`:`Config` 新增 `AdminOpenids []string` 字段,从 `ADMIN_OPENIDS`(逗号分隔)解析,去空白与空项

## 2. Service 更新方法

- [x] 2.1 `internal/appconfig/service.go`:新增 `UpdateAuditMode(ctx, auditMode bool) error`,更新单行 `audit_mode`(不存在则创建单行)

## 3. 管理员白名单中间件

- [x] 3.1 新增管理员鉴权中间件:从 `c.GetUint("userID")` 取 id,反查 `users.openid`,比对白名单;不在白名单返回 403,查不到用户返回 403
- [x] 3.2 中间件构造为可注入 db 与白名单集合的 `gin.HandlerFunc`,便于测试

## 4. Handler 与路由装配

- [x] 4.1 `internal/appconfig/handler.go`:新增 `UpdateConfig` PUT handler,解析 `auditMode` 请求体并调用 service
- [x] 4.2 `cmd/server/main.go`:在 `authorized` 分组下注册 `PUT /app-config`,前置管理员中间件

## 5. 单测

- [x] 5.1 中间件单测:白名单命中放行
- [x] 5.2 中间件单测:非白名单用户返回 403
- [x] 5.3 service 单测:`UpdateAuditMode` 更新后 `GetConfig` 返回新值

## 6. 验证

- [x] 6.1 运行 `gofmt`/项目格式化,`go build ./...` 与 `go test ./internal/appconfig/... ./pkg/...` 全绿
