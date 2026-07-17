# Implementation Tasks — diary-backend（后端私密日记）

## 1. Diary 模型

- [x] 1.1 `internal/diary/model.go`：Diary（Title/Content/Cover/Mood/Weather/AuthorID + Images 子表）、DiaryImage
- [x] 1.2 TableName、DTO（DiaryResponse/DiaryListItem/CreateDiaryInput/UpdateDiaryInput）
- [x] 1.3 toDiaryResponse / toListItem 映射（contentPreview 截前 100 字）

## 2. Service 层（私密强制）

- [x] 2.1 `internal/diary/service.go`：DiaryService 依赖 *gorm.DB
- [x] 2.2 `Create(ctx, userID, input)`：事务建 Diary + DiaryImage，返回详情
- [x] 2.3 `FindMine(ctx, userID, page, limit)`：仅本人，created_at DESC 分页
- [x] 2.4 `FindOne(ctx, id, userID)`：校验归属，不属于返回 ErrRecordNotFound
- [x] 2.5 `Update(ctx, id, userID, input)`：校验归属，更新字段 + 重建图片
- [x] 2.6 `Remove(ctx, id, userID)`：校验归属，删除（级联图片）

## 3. Handler 层

- [x] 3.1 `internal/diary/handler.go`：DiaryHandler + iface + NewDiaryHandler
- [x] 3.2 Create/List/Get/Update/Delete handler，JWT 取 userID，参数解析
- [x] 3.3 归属校验失败映射 404，复用 response 包

## 4. 注册路由 + Migrate

- [x] 4.1 `cmd/server/main.go`：authorized 组注册 5 个 diary 路由
- [x] 4.2 AutoMigrate 加 Diary/DiaryImage

## 5. 测试

- [x] 5.1 `internal/diary/handler_test.go`：mock service，测 CRUD + 未授权 + 归属校验
- [x] 5.2 `go build ./... && go test ./...` 通过
