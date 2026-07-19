# Go 学习大纲 — 前端开发者转型指南

> 面向有前端（JavaScript/TypeScript）经验的开发者，系统学习 Go 语言的完整路径。
> 从语言基础到框架应用，从生态工具到实战项目，循序渐进。

---

## 阶段一：Go 语言基础（2-3 周）

### 1.1 环境搭建与工具链

- [x] 安装 Go（官方安装 / Homebrew）
- [x] 理解 `GOPATH` vs Go Modules（`go.mod`）
- [x] 编辑器配置（VS Code + Go 插件 / GoLand）
- [x] 常用命令：`go run`、`go build`、`go test`、`go mod tidy`
- [x] 对比前端：Go Modules ≈ package.json + node_modules

### 1.2 基本语法

- [x] 变量声明：`var`、`:=`（短变量声明）
- [x] 基本类型：`int`、`float64`、`string`、`bool`、`byte`、`rune`
- [x] 常量与 `iota` 枚举
- [x] 控制流：`if`、`for`（Go 只有 for 循环）、`switch`
- [x] 对比前端：没有 `while`、没有 `class`、没有 `this`

### 1.3 函数

- [x] 函数定义与多返回值
- [x] 命名返回值
- [x] 可变参数 `...`
- [x] 匿名函数与闭包
- [x] `defer` 延迟执行（类似 finally）
- [x] 对比前端：Go 函数是一等公民，但没有 `=>` 箭头函数语法

### 1.4 复合类型

- [x] 数组与切片（Slice）— 对比 JS 的 Array
- [x] Map — 对比 JS 的 Object/Map
- [x] 结构体（Struct）— 对比 JS 的 Class/Object
- [x] 指针基础：`&` 取地址、`*` 解引用
- [x] 值类型 vs 引用类型

### 1.5 面向对象（Go 风格）

- [x] 方法（Method）：给 struct 绑定方法
- [x] 接口（Interface）：隐式实现（鸭子类型）
- [ ] 组合优于继承：struct 嵌入
- [ ] 空接口 `interface{}` / `any`
- [ ] 类型断言与类型 switch
- [ ] 对比前端：没有 class 继承链，接口无需 `implements`

### 1.6 错误处理

- [x] `error` 接口与错误返回模式
- [x] `errors.New()` 与 `fmt.Errorf()`
- [ ] 自定义错误类型
- [ ] `errors.Is()` / `errors.As()` 错误判断
- [x] `panic` / `recover`（极少使用）
- [ ] 对比前端：没有 try/catch，错误是返回值

---

## 阶段二：Go 进阶特性（2-3 周）

### 2.1 并发编程（Go 核心优势）

- [ ] Goroutine：轻量级线程（`go func()`）
- [ ] Channel：goroutine 间通信
- [ ] 无缓冲 channel vs 有缓冲 channel
- [ ] `select` 多路复用
- [ ] `sync.WaitGroup`：等待一组 goroutine 完成
- [ ] `sync.Mutex` / `sync.RWMutex`：互斥锁
- [ ] `context.Context`：超时、取消、传值
- [ ] 对比前端：goroutine ≈ Web Worker，但更轻量；channel ≈ 消息通道

### 2.2 泛型（Go 1.18+）

- [ ] 类型参数基础语法
- [ ] 类型约束（constraints）
- [ ] 泛型函数与泛型结构体
- [ ] 常用约束：`comparable`、`any`
- [ ] 对比前端：类似 TypeScript 的泛型 `<T>`

### 2.3 包管理与项目结构

- [ ] `go.mod` 与 `go.sum`
- [ ] 包的导入与可见性（大写导出）
- [ ] `internal/` 包的访问限制
- [ ] 标准项目布局（参考本项目）：
  ```
  cmd/          → 入口点
  internal/     → 私有业务逻辑
  pkg/          → 可导出的公共库
  docs/         → 文档
  ```
- [ ] 对比前端：大写导出 ≈ `export`，internal ≈ 私有模块

### 2.4 测试

- [ ] `testing` 包：`TestXxx` 函数
- [ ] 表驱动测试（Table-Driven Tests）
- [ ] `go test ./...` 运行测试
- [ ] 基准测试 `BenchmarkXxx`
- [ ] 测试覆盖率 `go test -cover`
- [ ] Mock 与接口测试
- [ ] 对比前端：内置测试框架，无需 Jest/Vitest

### 2.5 标准库精选

- [ ] `fmt` — 格式化输出
- [ ] `net/http` — HTTP 客户端与服务端
- [ ] `encoding/json` — JSON 序列化/反序列化（struct tag）
- [ ] `io` / `os` — 文件与 IO 操作
- [ ] `time` — 时间处理
- [ ] `strconv` — 字符串转换
- [ ] `log` / `slog` — 日志
- [ ] `strings` / `bytes` — 字符串/字节操作

---

## 阶段三：Web 开发框架（2-3 周）

### 3.1 Gin 框架（本项目使用）

- [ ] 路由注册：`GET`、`POST`、`PUT`、`DELETE`
- [ ] 路由分组 `Group`
- [ ] 路径参数与查询参数
- [ ] 请求体绑定：`ShouldBindJSON`
- [ ] 响应：`c.JSON()`、状态码
- [ ] 中间件机制：`Use()`、`Next()`、`Abort()`
- [ ] 自定义中间件：日志、认证、CORS、错误恢复
- [ ] 对比前端：Gin ≈ Express.js，中间件模型几乎一致

### 3.2 GORM（ORM 框架）

- [ ] 数据库连接配置
- [ ] 模型定义与 struct tag
- [ ] 自动迁移 `AutoMigrate`
- [ ] CRUD 操作：`Create`、`First`、`Find`、`Save`、`Delete`
- [ ] 查询条件：`Where`、`Or`、`Order`、`Limit`、`Offset`
- [ ] 关联关系：`HasMany`、`BelongsTo`、`ManyToMany`
- [ ] 预加载 `Preload`
- [ ] 事务处理
- [ ] 钩子函数 `BeforeCreate`、`AfterUpdate` 等
- [ ] 对比前端：GORM ≈ Prisma/TypeORM

### 3.3 认证与安全

- [ ] JWT 原理与实现（`golang-jwt`）
- [ ] 中间件鉴权
- [ ] 密码哈希：`bcrypt`
- [ ] 请求验证：`validator` tag
- [ ] CORS 配置

### 3.4 其他 Web 框架了解

- [ ] 标准库 `net/http`（原生开发）
- [ ] Echo — 高性能框架
- [ ] Fiber — 类 Express 风格
- [ ] Chi — 轻量路由
- [ ] Hertz — 字节跳动出品

---

## 阶段四：Go 生态工具链（1-2 周）

### 4.1 开发工具

- [ ] `air` — 热重载（类似 nodemon）
- [ ] `golangci-lint` — 代码检查（类似 ESLint）
- [ ] `gofmt` / `goimports` — 代码格式化（类似 Prettier）
- [ ] `delve` — 调试器
- [ ] `go vet` — 静态分析
- [ ] `go generate` — 代码生成

### 4.2 API 文档

- [ ] Swagger/OpenAPI：`swaggo/swag`
- [ ] 注释驱动生成文档
- [ ] Swagger UI 集成

### 4.3 配置管理

- [ ] `godotenv` — .env 文件加载
- [ ] `viper` — 强大的配置管理库
- [ ] 环境变量 `os.Getenv`

### 4.4 日志

- [ ] 标准库 `log` / `slog`（Go 1.21+）
- [ ] `zap`（Uber 出品，高性能）
- [ ] `logrus`（结构化日志）

### 4.5 数据库与缓存

- [ ] MySQL / PostgreSQL 驱动
- [ ] Redis：`go-redis`
- [ ] MongoDB：`mongo-go-driver`
- [ ] 数据库迁移：`golang-migrate`

### 4.6 消息队列与任务

- [ ] RabbitMQ / Kafka 客户端
- [ ] `asynq` — 异步任务队列（类似 BullMQ）
- [ ] `cron` — 定时任务

### 4.7 HTTP 客户端

- [ ] 标准库 `net/http`
- [ ] `resty` — 更友好的 HTTP 客户端（类似 axios）

---

## 阶段五：微服务与高级架构（2-3 周）

### 5.1 微服务基础

- [ ] gRPC 与 Protocol Buffers
- [ ] RESTful vs gRPC 选择
- [ ] 服务注册与发现
- [ ] API Gateway

### 5.2 微服务框架

- [ ] Go-kit — 微服务工具集
- [ ] Go-micro — 微服务框架
- [ ] Kratos — B站开源框架
- [ ] Go-Zero — 国产微服务框架（好未来）

### 5.3 容器化与部署

- [ ] Dockerfile 编写（多阶段构建）
- [ ] Docker Compose 编排
- [ ] Kubernetes 基础部署
- [ ] CI/CD 集成（GitHub Actions）
- [ ] 对比前端：Go 编译为单一二进制，部署极简

### 5.4 可观测性

- [ ] 链路追踪：OpenTelemetry
- [ ] 指标监控：Prometheus + Grafana
- [ ] 日志聚合：ELK / Loki
- [ ] 健康检查接口

### 5.5 性能优化

- [ ] `pprof` 性能分析
- [ ] 内存逃逸分析
- [ ] sync.Pool 对象池
- [ ] 连接池管理

---

## 阶段六：实战项目（持续）

### 6.1 初级项目

- [ ] CLI 命令行工具（TODO 管理器）
- [ ] RESTful API 服务（博客系统）
- [ ] 文件上传服务

### 6.2 中级项目（参考本项目 go-service）

- [ ] 完整的微信小程序后端
  - 用户认证（JWT + 微信登录）
  - 帖子 CRUD
  - 互动系统（点赞、收藏、评论）
  - 文件上传（云存储）
  - Swagger 文档
  - Docker 部署

### 6.3 高级项目

- [ ] 实时聊天系统（WebSocket + Redis Pub/Sub）
- [ ] 分布式爬虫系统
- [ ] API Gateway 网关服务
- [ ] 秒杀/抢购系统（并发控制）

---

## 学习资源推荐

### 官方资源

| 资源 | 链接 | 说明 |
|------|------|------|
| Go 官网 | https://go.dev | 官方文档与教程 |
| Go Tour | https://go.dev/tour | 交互式入门教程 |
| Effective Go | https://go.dev/doc/effective_go | 官方最佳实践 |
| Go Blog | https://go.dev/blog | 官方技术博客 |
| Go Playground | https://go.dev/play | 在线运行代码 |

### 书籍

| 书名 | 适合阶段 |
|------|----------|
| 《Go 程序设计语言》（The Go Programming Language） | 基础-进阶 |
| 《Go 语言实战》（Go in Action） | 基础-进阶 |
| 《Go 语言高级编程》 | 进阶-高级 |
| 《Go 语言设计与实现》 | 高级（源码级） |

### 在线课程/教程

| 资源 | 说明 |
|------|------|
| [Go by Example](https://gobyexample.com) | 示例驱动学习 |
| [Learn Go with Tests](https://quii.gitbook.io/learn-go-with-tests) | TDD 方式学习 |
| [Gopher Academy Blog](https://blog.gopheracademy.com) | 社区博客 |

---

## 前端 → Go 概念对照表

| 前端概念 | Go 对应概念 | 说明 |
|----------|-------------|------|
| `package.json` | `go.mod` | 依赖管理 |
| `node_modules` | `$GOPATH/pkg/mod` | 依赖缓存 |
| `npm install` | `go mod tidy` | 安装/整理依赖 |
| `npm run dev` | `go run .` / `air` | 运行项目 |
| `npm run build` | `go build` | 构建产物 |
| `export` | 首字母大写 | 导出可见性 |
| `import` | `import` | 包导入 |
| `class` | `struct` + 方法 | 面向对象 |
| `interface` | `interface`（隐式） | 接口定义 |
| `async/await` | goroutine + channel | 异步/并发 |
| `Promise.all` | `sync.WaitGroup` | 等待多个异步 |
| `try/catch` | `if err != nil` | 错误处理 |
| `Express.js` | Gin / Echo | Web 框架 |
| `Prisma/TypeORM` | GORM / sqlx | 数据库 ORM |
| `Jest/Vitest` | `testing` 标准包 | 测试框架 |
| `ESLint` | `golangci-lint` | 代码检查 |
| `Prettier` | `gofmt` | 代码格式化 |
| `nodemon` | `air` | 热重载 |
| `axios` | `net/http` / `resty` | HTTP 客户端 |
| `dotenv` | `godotenv` | 环境变量 |
| `TypeScript 泛型` | Go 泛型（1.18+） | 类型参数化 |

---

## 学习建议

1. **不要用前端思维写 Go** — Go 追求简单直接，没有魔法语法
2. **拥抱错误处理** — `if err != nil` 虽然冗长，但让错误路径显式化
3. **善用 goroutine** — 这是 Go 最大的优势，但也要注意数据竞争
4. **多读标准库源码** — Go 的标准库是最好的学习材料
5. **先写测试** — Go 内置测试框架非常好用，养成 TDD 习惯
6. **参考本项目** — `go-service` 项目展示了真实的 Go Web 应用结构

---

*最后更新：2026-07-01*
