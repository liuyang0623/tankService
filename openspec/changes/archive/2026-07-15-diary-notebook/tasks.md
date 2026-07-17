# Implementation Tasks — diary-notebook（后端日记本）

## 1. Notebook 模型 + service

- [x] 1.1 `internal/notebook/model.go`：Notebook 模型 + DTO（NotebookResponse/CreateNotebookInput/UpdateNotebookInput）+ DefaultNotebookName
- [x] 1.2 `internal/notebook/service.go`：EnsureDefault/FindMine/Create/Update/Remove
- [x] 1.3 `internal/notebook/service_test.go`：常量 + 输入零值测试

## 2. Notebook handler + 路由

- [x] 2.1 `internal/notebook/handler.go`：Create/FindMine/Update/Remove + iface
- [x] 2.2 `internal/notebook/handler_test.go`：mock service，CRUD + 未授权 + 404
- [x] 2.3 `cmd/server/main.go`：注册 4 路由 + AutoMigrate Notebook

## 3. diary 归属 notebook

- [x] 3.1 `internal/diary/model.go`：Diary 加 NotebookID，DTO 加 notebookId
- [x] 3.2 `internal/diary/service.go`：Create 落 NotebookID，FindMine 加过滤，映射填 notebookId
- [x] 3.3 `internal/diary/handler.go`：FindMine 读 notebookId query

## 4. 验证

- [x] 4.1 `go build ./... && go test ./...` 通过
