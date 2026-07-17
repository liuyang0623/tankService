---
comet_change: go-service
role: technical-design
canonical_spec: openspec
---

## Context

构建一个面向微信小程序的 Go 后端服务，提供用户认证、内容管理、社交互动和文件上传功能。采用 Gin + GORM + MySQL 技术栈，按功能模块分包，分模块逐步实现，每个模块独立可验证。

## Goals / Non-Goals

**Goals:**
- 实现完整的 REST API，前缀 `/api/v1`
- 按功能模块组织代码（`internal/<module>/`）
- GORM AutoMigrate 自动建表
- JWT Bearer Token 认证
- Swagger 文档自动生成
- 统一响应格式与错误处理

**Non-Goals:**
- 不编写自动化测试（可后续补充）
- 不配置 CI/CD 或 Docker（可后续补充）
- 不实现管理后台

## Decisions

### 1. 目录结构：按功能模块分包

```
go-service/
├── cmd/
│   └── server/
│       └── main.go          # 程序入口
├── internal/
│   ├── auth/                # 认证模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── model.go
│   ├── users/               # 用户模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── model.go
│   ├── posts/               # 文章模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── model.go
│   ├── interactions/        # 互动模块
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── model.go
│   └── upload/              # 上传模块
│       ├── handler.go
│       └── service.go
├── pkg/
│   ├── config/              # 配置加载（godotenv）
│   ├── database/            # GORM 数据库连接
│   ├── middleware/          # JWT 中间件、CORS
│   └── response/            # 统一响应格式
├── docs/                    # swaggo 生成的 Swagger 文档
├── .env                     # 环境变量
├── go.mod
└── go.sum
```

**理由**：每个功能模块自包含（handler + service + model），便于独立开发和验证。

### 2. 认证：JWT Bearer Token

- 微信登录：客户端发送 `code`，服务端调用微信 `jscode2session` 接口换取 `openid`，查找或创建用户，签发 JWT
- JWT 中间件：从 `Authorization: Bearer <token>` 提取 token，验证后将 `userID` 注入 Gin context
- 使用 `golang-jwt/jwt/v5`

### 3. ORM：GORM + AutoMigrate

- 所有数据模型在 `internal/<module>/model.go` 中定义
- 启动时统一调用 `db.AutoMigrate(...)` 同步表结构
- 表名与原 Prisma schema 保持一致（`users`, `posts`, `post_images` 等）

### 4. Swagger：swaggo/swag

- 使用注释驱动生成（`// @Summary`、`// @Param` 等）
- 文档路径：`/api/docs`
- 需在开发时运行 `swag init` 生成 `docs/` 目录

### 5. 统一响应格式

```go
// 成功
{"data": <payload>, "message": "success"}
// 失败
{"error": "错误描述", "code": <http_status>}
```

### 6. 又拍云上传：HMAC-SHA1 签名

签名算法：
1. 对密码做 MD5，得到 `password_md5`
2. 构造签名串：`METHOD&/bucket/uri&date[&content_md5]`
3. 用 `password_md5` 做 HMAC-SHA1，结果 Base64 编码
4. `Authorization: UPYUN operator:signature`

### 7. 分模块实现顺序

1. **project-bootstrap**：项目骨架、配置、数据库、路由、Swagger
2. **auth**：微信登录、JWT 中间件
3. **users**：用户 CRUD、个人资料
4. **posts**：文章 CRUD、话题、图片、分页
5. **interactions**：点赞、收藏、评论
6. **upload**：又拍云上传

每个模块完成后可独立启动服务验证。

## Risks / Trade-offs

- **微信接口调用**：`jscode2session` 需要真实的 AppID/Secret，开发阶段可用 mock 测试
  → 缓解：配置读自 `.env`，不硬编码
- **GORM AutoMigrate 限制**：不会删除废弃列，列类型变更需手动处理
  → 缓解：开发阶段可接受，生产环境再补 migration
- **swag init 需要手动执行**：Swagger 文档不会自动更新
  → 缓解：在 `README` 中注明，或加入 Makefile

## Open Questions

- 无，所有关键决策已确认
