# Design: profile 接口用户不存在时返回 401

## 修复方案

在 `internal/users/handler.go` 中，将「凭当前登录 userID 查自身」的接口在
`gorm.ErrRecordNotFound` 时的返回从 `404` 改为 `401`：

- `GetProfile`：`err == gorm.ErrRecordNotFound` 分支改为 `response.Unauthorized(c, "login expired")`。
- `UpdateProfile`：同一分支改为 `response.Unauthorized(c, "login expired")`。

`response.Unauthorized(c, msg)` 内部写 `401` + `{"data":null,"code":401,"message":msg}`，
与项目统一响应契约一致；前端据 401 触发重新登录。

`GetUser`（`GET /users/:id`，按路径 ID 查他人）**保持 404 不变**——那是查询他人资源，
资源不存在返回 404 是正确语义。

## 为何是 401 而非 404

profile 类接口的主体是「当前登录用户自身」，其 userID 来自已通过签名校验的 JWT。
若该用户在库中不存在，说明这枚 token 不再对应有效账号（跨环境旧 token、账号已删除等），
属于登录态失效，应返回 401 让客户端重新走登录流程；返回 404 会让前端误判为业务资源缺失。

## 不做的事

- 不改 `JWTMiddleware`（签名校验逻辑正确，token 本身有效）。
- 不改 service 层（service 如实返回 `gorm.ErrRecordNotFound`，语义映射属 handler 职责）。
- 不改 `GetUser` 的 404 行为。

## 验证方式

- handler 单测：mock service 返回 `gorm.ErrRecordNotFound`，断言 `GetProfile` /
  `UpdateProfile` 返回状态码 401。
- `go build ./...` 与 `go test ./...` 全绿。
