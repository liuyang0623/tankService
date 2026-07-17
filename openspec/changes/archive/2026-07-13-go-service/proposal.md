## Why

构建 go-service：一个面向微信小程序的后端 Go 服务，提供用户认证、内容发布、社交互动和文件上传功能。采用 Go + Gin + GORM + MySQL 技术栈，以模块化方式逐步实现各功能模块。

## What Changes

- 新建 Go 项目骨架（Gin HTTP 框架 + GORM ORM + MySQL）
- 实现微信小程序登录及 JWT 认证模块
- 实现用户管理模块（个人资料 CRUD）
- 实现内容模块（文章发布/草稿、话题标签、图片列表、分页）
- 实现社交互动模块（点赞/取消点赞、收藏/取消收藏、评论及嵌套回复）
- 实现文件上传模块（又拍云 Upyun，HMAC-SHA1 签名认证）
- 集成 Swagger 文档（swaggo/swag）
- 统一错误处理与响应格式

## Capabilities

### New Capabilities

- `project-bootstrap`: Go 项目初始化，包括目录结构、依赖管理、配置加载、数据库连接、路由注册、Swagger 集成
- `auth`: 微信小程序 code → openid → JWT 登录流程，JWT 中间件
- `users`: 用户 CRUD、个人资料查询与更新
- `posts`: 文章 CRUD（草稿/发布状态切换）、话题标签（Topic）关联、图片列表、分页列表
- `interactions`: 文章点赞/取消点赞、收藏/取消收藏、评论（含嵌套回复）、用户收藏列表
- `upload`: 又拍云文件上传（图片/通用文件），HMAC-SHA1 签名生成

### Modified Capabilities

（无，这是全新项目）

## Impact

- **新增**：`/Users/liuyang/mywork/refactor/go-service/` 下完整 Go 项目代码
- **API 前缀**：`/api/v1`
- **依赖**：gin, gorm, gorm/driver/mysql, golang-jwt/jwt, swaggo/swag, godotenv, go-resty/resty（HTTP 客户端）
- **配置**：通过 `.env` 配置 DATABASE_URL、WECHAT_APPID、WECHAT_SECRET、JWT_SECRET、UPYUN_* 等环境变量
- **数据库**：GORM AutoMigrate 自动同步表结构（User, Topic, Post, PostImage, PostTopic, Like, Favorite, Comment）
