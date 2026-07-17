## Why

首页正在从"单一信息流"改造为"搜索 + 分类 tab"的内容中心（前端 change home-revamp）。前端需要后端提供：固定分类列表、按分类/关键词/排序/关注关系筛选文章的能力，以及发布时写入分类。本 change 交付这些后端能力，是前端改造的前置依赖。

用户已确认的设计决策：
- 分类固定 5 类（故事/日常/技术/美食/旅游），后端用 Post 的枚举字符串字段 `category` 存储
- `GET /categories` 返回固定分类列表（不支持后台增删）
- `GET /posts` 扩展查询参数：`keyword`（title 模糊查）、`category`（分类过滤）、`sort=likes`（推荐排序）、`following=true`（关注流）
- 发布/更新接口接受 `category` 字段

## What Changes

- **Post 模型**：新增 `Category` 字段（`varchar(20)`，可空，兼容旧数据）
- **分类接口**：`GET /categories` 返回固定分类列表（value + label）
- **文章列表查询扩展**（`GET /posts`）：
  - `keyword=xxx`：按 title `LIKE %xxx%` 模糊查
  - `category=story`：按分类精确过滤
  - `sort=likes`：按 `like_count` 倒序（推荐流），默认仍按 `published_at` 倒序
  - `following=true`：仅返回当前登录用户关注的作者的文章（需鉴权；未登录返回空或忽略）
- **发布/更新**：`CreatePostInput` / `UpdatePostInput` 增加 `Category` 字段
- **响应**：`PostResponse` / `PostListResponse` 返回 `category`

## Capabilities

### New Capabilities
- `post-category`: 文章分类能力——固定分类字段、分类列表接口、发布写入分类

### Modified Capabilities
- `posts`: 文章列表查询扩展 keyword/category/sort/following 参数

## Impact

- **模型变更**：`internal/posts/model.go` Post 加 `Category string`，AutoMigrate 自动加列
- **新增**：`GET /categories` handler + service 方法（固定列表，无 DB 查询）
- **修改**：
  - `internal/posts/service.go`：`FindAll` 签名扩展查询选项（keyword/category/sort/following/currentUserID）
  - `internal/posts/handler.go`：`FindAll` 解析新 query 参数
  - `CreatePostInput`/`UpdatePostInput`/`PostResponse`/`PostListResponse` 加 category
  - `cmd/server/main.go`：注册 `GET /categories` 路由
- **关注流**：join follows 表（follow 模块已有 Follow 模型），需要当前用户 id
- **前端**：本 change 不含前端；前端在 change② home-revamp 消费这些接口
- **分类固定**：category 合法值在后端常量维护，不建独立 categories 表（YAGNI）
- **旧数据兼容**：无 category 的文章归入前端"其他" tab（category 为空）
