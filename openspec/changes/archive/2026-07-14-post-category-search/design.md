## Context

go-service 现有 posts 模块：`Post` 模型含 Title/Content/Cover/Status/AuthorID/LikeCount 等，`FindAll(page, limit)` 只按 `published_at DESC` 查已发布文章。已有 `Topic` 表 + `post_topics` 多对多（自由 #话题，与本次固定分类是不同语义，保持不变）。follow 模块已有 `Follow{FollowerID, FollowingID}` 模型。

现状要点：
- `FindAll` 查询链：`Where(status=PUBLISHED).Preload(Author/Images/Topics).Order(published_at DESC)` — 扩展点清晰
- `PostResponse`/`PostListResponse` 是 API 响应 DTO，需加 category 字段
- 互动模块已有 like_count 冗余在 Post 上（推荐排序直接用，无需 join）
- follow 关注关系查询：`SELECT following_id FROM follows WHERE follower_id = ?`

## Goals / Non-Goals

**Goals:**
- Post 加 category 枚举字符串字段
- `GET /categories` 固定分类列表接口
- `GET /posts` 扩展 keyword/category/sort/following 查询
- 发布/更新写入 category

**Non-Goals:**
- 不建独立 categories 表（固定 5 类，后端常量维护）
- 不改自由 #话题（Topic）机制
- 不做全文搜索（仅 title LIKE）
- 不做分类的后台增删管理
- 不改前端（change② 负责）

## Decisions

### D1. category 存为枚举字符串
- Post 加 `Category string \`gorm:"type:varchar(20);index"\``，可空
- 合法值常量：`story`/`daily`/`tech`/`food`/`travel`
- 空值兼容旧数据（前端归"其他"）

### D2. GET /categories 返回固定列表
- 无 DB 查询，返回后端常量：`[{value:"story", label:"故事"}, ...]`
- 公开接口（无需鉴权）

### D3. FindAll 查询扩展
- 签名扩展为接受 options struct：`{keyword, category, sort, following, currentUserID}`
- `keyword` → `Where("title LIKE ?", "%"+kw+"%")`
- `category` → `Where("category = ?", cat)`
- `sort=likes` → `Order("like_count DESC")`，否则 `Order("published_at DESC")`
- `following=true` + currentUserID → `Where("author_id IN (SELECT following_id FROM follows WHERE follower_id = ?)", uid)`
- 多条件可组合（category + keyword 等）

### D4. handler 解析 query
- `FindAll` 用 OptionalJWTMiddleware（following 需要 uid，未登录时 following 忽略/返回空）
- 解析 `c.Query("keyword")`/`category`/`sort`/`following`

### D5. 发布/更新写 category
- `CreatePostInput`/`UpdatePostInput` 加 `Category` 字段
- Create/Update service 写入；校验 category 合法值（非法则忽略或报错）

## Risks / Open Questions

- following=true 未登录：返回空列表（前端未登录隐藏关注 tab，不会触发）
- category 非法值：Create 时校验，非法返回 400
- 推荐排序用冗余 like_count，实时性足够（点赞已更新该字段）
