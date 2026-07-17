# tankService

微信小程序后端服务，基于 Go + Gin + GORM + MySQL 构建，提供用户、帖子、互动、社交、私信、日记、系统通知、文件上传等 RESTful API。

> AI 协作前置条件与铁律见 [AGENT.md](./AGENT.md)——本仓库用 **comet + OpenSpec** 管理开发全流程，勿裸开发。

## 技术栈

- **语言**：Go 1.25
- **Web 框架**：[Gin](https://github.com/gin-gonic/gin)
- **ORM**：[GORM](https://gorm.io) + MySQL 8.0
- **认证**：JWT + 微信小程序登录
- **实时**：WebSocket（私信）
- **文件上传**：又拍云（UpYun）
- **API 文档**：Swagger（swaggo）
- **容器化**：Docker / Docker Compose
- **开发流程**：comet + OpenSpec（spec-driven）

## 目录结构

```
.
├── cmd/server/          # 应用入口 + 路由装配 + AutoMigrate
├── internal/
│   ├── auth/            # 微信登录与 JWT 认证
│   ├── users/           # 用户模块
│   ├── posts/           # 帖子模块（含话题分类）
│   ├── interactions/    # 互动模块（点赞、收藏、评论）
│   ├── follow/          # 关注/粉丝
│   ├── message/         # 私信会话（WebSocket）
│   ├── notification/    # 系统通知（关注等，可扩展 like/comment）
│   ├── diary/           # 日记
│   ├── notebook/        # 日记本
│   ├── wechat/          # 微信服务端能力（access_token 缓存 + 订阅消息发送）
│   ├── subscribepush/   # 关注事件 → 微信订阅消息推送（配额管理）
│   └── upload/          # 文件上传
├── pkg/                 # response/config/database/middleware 等公共包
├── openspec/            # 需求规格与 change 归档（comet 工作流）
├── docs/                # Swagger 自动生成文档
├── docker-compose.yml       # 生产环境
├── docker-compose.dev.yml   # 本地开发（仅数据库）
└── .env.example         # 环境变量模板
```

## 快速开始

### 前提条件

- Go 1.25+
- Docker & Docker Compose
- Node.js（用于运行 npm scripts，可选）

### 1. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env，填入真实的数据库密码、JWT 密钥、微信 AppID/Secret、又拍云配置等
```

### 2. 本地开发（推荐）

启动 MySQL 数据库容器，然后用 `go run` 直接运行服务：

```bash
# 启动开发数据库
npm run dev:db

# 启动 Go 服务（热重载可配合 air 使用）
npm run dev
```

服务默认运行在 `http://localhost:3000`，API 路径前缀为 `/api/v1`。

### 3. Docker 完整部署

```bash
# 启动所有服务（MySQL + Go App）
npm run start

# 停止服务
npm run stop
```

## 可用命令

| 命令 | 说明 |
|------|------|
| `npm run dev` | 本地直接运行服务（`go run`） |
| `npm run dev:db` | 启动开发用 MySQL 容器 |
| `npm run dev:db:down` | 停止开发数据库容器 |
| `npm run dev:db:logs` | 查看开发数据库日志 |
| `npm run start` | Docker Compose 启动完整服务栈 |
| `npm run stop` | 停止完整服务栈 |
| `npm run build` | 编译为二进制文件 `./server` |
| `npm run test` | 运行所有测试 |
| `npm run test:v` | 运行测试（详细输出） |
| `npm run swagger` | 重新生成 Swagger 文档 |
| `npm run tidy` | 整理 Go 模块依赖 |

## 环境变量说明

| 变量 | 说明 | 示例 |
|------|------|------|
| `DATABASE_URL` | MySQL 连接字符串 | `root:pwd@tcp(127.0.0.1:3307)/go_service_db?...` |
| `DB_ROOT_PASSWORD` | MySQL root 密码 | `your_root_password` |
| `DB_NAME` | 数据库名称 | `go_service_db` |
| `DB_PORT` | 宿主机映射端口 | `3307` |
| `JWT_SECRET` | JWT 签名密钥 | `your_jwt_secret` |
| `WECHAT_APPID` | 微信小程序 AppID | `wx...` |
| `WECHAT_SECRET` | 微信小程序 Secret | `...` |
| `WECHAT_SUBSCRIBE_TPL_FOLLOW` | 关注订阅消息模板 ID | `Q2Bce...T_BA` |
| `UPYUN_BUCKET` | 又拍云空间名 | `my-bucket` |
| `UPYUN_DOMAIN` | 又拍云访问域名 | `https://img.example.com` |
| `UPYUN_OPERATOR` | 又拍云操作员 | `operator_name` |
| `UPYUN_PASSWORD` | 又拍云操作员密码 | `...` |
| `PORT` | 服务监听端口 | `3000` |
| `API_PREFIX` | API 路径前缀 | `api/v1` |
| `CORS_ORIGIN` | 允许的跨域来源 | `http://localhost:3000` |

## API 文档

服务启动后访问 Swagger UI：

```
http://localhost:3000/api/docs/index.html
```

## 主要接口

| 方法 | 路径 | 说明 | 是否需要 JWT |
|------|------|------|:---:|
| `POST` | `/api/v1/auth/wechat/login` | 微信小程序登录 | ✗ |
| `GET` | `/api/v1/users/:id` | 获取用户信息 | ✗ |
| `GET` | `/api/v1/users/profile` | 获取当前用户资料 | ✓ |
| `PATCH` | `/api/v1/users/profile` | 更新用户资料 | ✓ |
| `POST` | `/api/v1/users/subscribe/follow` | 上报关注订阅授权（累加推送配额） | ✓ |
| `GET` | `/api/v1/posts` | 获取帖子列表 | ✗ |
| `GET` | `/api/v1/posts/:id` | 获取帖子详情 | ✗ |
| `POST` | `/api/v1/posts` | 创建帖子 | ✓ |
| `PATCH` | `/api/v1/posts/:id` | 更新帖子 | ✓ |
| `DELETE` | `/api/v1/posts/:id` | 删除帖子 | ✓ |
| `POST` | `/api/v1/posts/:id/publish` | 发布帖子 | ✓ |
| `POST` | `/api/v1/posts/:id/like` | 点赞帖子 | ✓ |
| `POST` | `/api/v1/posts/:id/favorite` | 收藏帖子 | ✓ |
| `GET` | `/api/v1/posts/:id/comments` | 获取帖子评论 | ✗ |
| `POST` | `/api/v1/comments` | 发表评论 | ✓ |
| `DELETE` | `/api/v1/comments/:id` | 删除评论 | ✓ |
| `POST` | `/api/v1/upload/image` | 上传图片 | ✓ |
| `POST` | `/api/v1/upload/file` | 上传文件 | ✓ |
| `POST` | `/api/v1/users/:id/follow` | 关注/取关用户 | ✓ |
| `GET` | `/api/v1/users/:id/followers` | 粉丝列表 | 可选 |
| `GET` | `/api/v1/users/:id/following` | 关注列表 | 可选 |
| `GET` | `/api/v1/conversations` | 私信会话列表 | ✓ |
| `GET` | `/api/v1/conversations/:id/messages` | 会话消息 | ✓ |
| `POST` | `/api/v1/messages` | 发送私信 | ✓ |
| `POST` | `/api/v1/conversations/:id/read` | 会话标记已读 | ✓ |
| `GET` | `/ws` | WebSocket（私信推送） | 内部校验 token |
| `GET` | `/api/v1/notifications` | 系统通知列表 | ✓ |
| `POST` | `/api/v1/notifications/read` | 通知整体标记已读 | ✓ |
| `GET` | `/api/v1/notifications/unread-count` | 未读数 + 最新摘要 | ✓ |
| `GET/POST/PATCH/DELETE` | `/api/v1/diaries`、`/api/v1/notebooks` | 日记 / 日记本 CRUD | ✓ |
