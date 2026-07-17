# project-bootstrap Specification

## Purpose
TBD - created by archiving change go-service. Update Purpose after archive.
## Requirements
### Requirement: 项目初始化与配置加载
系统 SHALL 在启动时从 `.env` 文件加载配置，包括数据库连接字符串、JWT 密钥、微信配置、又拍云配置。

#### Scenario: 配置加载成功
- **WHEN** 服务启动且 `.env` 文件存在
- **THEN** 所有配置项被加载，服务正常启动

#### Scenario: 配置缺失
- **WHEN** 必填配置项（如 DATABASE_URL）缺失
- **THEN** 服务拒绝启动并输出错误信息

### Requirement: 数据库连接与 AutoMigrate
系统 SHALL 在启动时连接 MySQL 数据库，并通过 GORM AutoMigrate 同步所有数据模型的表结构。

#### Scenario: 数据库连接成功
- **WHEN** DATABASE_URL 有效且数据库可达
- **THEN** GORM 成功连接，AutoMigrate 同步表结构

### Requirement: HTTP 服务与路由注册
系统 SHALL 使用 Gin 框架提供 HTTP 服务，所有 API 路由使用 `/api/v1` 前缀，并启用 CORS。

#### Scenario: 服务启动
- **WHEN** 服务启动成功
- **THEN** 监听配置的端口（默认 3000），输出启动日志

#### Scenario: CORS
- **WHEN** 跨域请求到达
- **THEN** 响应包含正确的 CORS 头

### Requirement: Swagger 文档
系统 SHALL 在 `/api/docs` 提供 Swagger UI 文档页面。

#### Scenario: 访问 Swagger
- **WHEN** GET `/api/docs/index.html`
- **THEN** 返回 Swagger UI 页面

