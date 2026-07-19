# Tasks: 修复 profile 接口用户不存在时返回 401

- [x] 1. 修改 `internal/users/handler.go` 的 `GetProfile`：`gorm.ErrRecordNotFound` 分支改为 `response.Unauthorized(c, "login expired")`
- [x] 2. 修改 `internal/users/handler.go` 的 `UpdateProfile`：`gorm.ErrRecordNotFound` 分支改为 `response.Unauthorized(c, "login expired")`
- [x] 3. 调整/补充 `internal/users/handler_test.go`：断言用户不存在时 `GetProfile`、`UpdateProfile` 返回 401
- [x] 4. 运行 `go build ./...` 与 `go test ./...`，确认全绿
