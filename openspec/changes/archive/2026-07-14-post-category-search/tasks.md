# Implementation Tasks — post-category-search（后端文章分类与搜索）

## 1. Post 模型加 category

- [x] 1.1 `internal/posts/model.go`：Post 加 `Category string \`gorm:"type:varchar(20);index"\``
- [x] 1.2 定义合法分类常量：story/daily/tech/food/travel + IsValidCategory 校验函数
- [x] 1.3 `cmd/server/main.go`：AutoMigrate 已含 Post，自动加列（确认）

## 2. 分类列表接口

- [x] 2.1 `internal/posts/service.go`：`ListCategories()` 返回固定列表 `[]CategoryInfo{value, label}`
- [x] 2.2 `internal/posts/handler.go`：`ListCategories` handler
- [x] 2.3 `cmd/server/main.go`：注册 `GET /categories`（公开）

## 3. FindAll 查询扩展

- [x] 3.1 定义 `FindAllOptions` struct：keyword/category/sort/following/currentUserID
- [x] 3.2 `FindAll` 签名改为接受 options，构建动态 Where/Order
- [x] 3.3 keyword → title LIKE；category → 精确过滤
- [x] 3.4 sort=likes → like_count DESC；默认 published_at DESC
- [x] 3.5 following + currentUserID → author_id IN (关注的人)
- [x] 3.6 `handler.go` FindAll 解析 query 参数 + 传 currentUserID（OptionalJWT）
- [x] 3.7 `main.go`：确认 `/posts` 用 OptionalJWTMiddleware（已是）

## 4. 响应加 category

- [x] 4.1 `PostResponse` / `PostListResponse` 加 `Category string` 字段
- [x] 4.2 `toPostResponse` / list 映射填充 category

## 5. 发布/更新写 category

- [x] 5.1 `CreatePostInput` / `UpdatePostInput` 加 `Category` 字段
- [x] 5.2 Create service 写入 + 校验合法值（非法返回错误）
- [x] 5.3 Update service 支持更新 category

## 6. 测试

- [x] 6.1 service 测试：FindAll 各查询组合（keyword/category/sort/following）
- [x] 6.2 category 校验测试（合法/非法值）
- [x] 6.3 ListCategories 返回固定列表测试
- [x] 6.4 `go build ./... && go test ./...` 通过
