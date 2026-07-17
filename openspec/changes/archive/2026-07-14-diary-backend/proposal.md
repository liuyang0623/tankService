## Why

移植 Moo 日记的日记功能（第一期：基础日记编辑），作为独立私密日记模块，与「摆烂随笔」公开社区分开。本 change 交付后端能力：日记表 + 私密 CRUD + 心情/天气字段。前端 change（diary-frontend）消费这些接口。

用户已确认的第一期范围：
- 富文本内容（content HTML）+ 封面/图片
- 心情 mood + 天气 weather 字段（前端 emoji 单选，后端存字符串）
- 私密：仅本人可见可管理（与 posts 的公开语义不同）

## What Changes

- **新增 diary 模块**（`internal/diary/`）：Diary 模型 + service + handler，照搬 posts 三件套模式
- **数据模型**：
  - `Diary`：Title/Content(HTML)/Cover/Mood/Weather/AuthorID + DiaryImage 子表（多图）
  - 私密：所有查询按 author_id 过滤，仅本人
- **REST 接口**（均需 JWT，且仅本人）：
  - `POST /diaries`：创建日记（title/content/cover/mood/weather/images）
  - `GET /diaries`：我的日记列表，分页，按创建时间倒序（时间线）
  - `GET /diaries/:id`：日记详情（校验归属）
  - `PATCH /diaries/:id`：更新（校验归属）
  - `DELETE /diaries/:id`：删除（校验归属）
- **图片**：复用现有 upload 接口拿 URL，diary 存 DiaryImage 关联
- **AutoMigrate**：新增 `&diary.Diary{}`、`&diary.DiaryImage{}`

## Capabilities

### New Capabilities
- `diary`: 后端私密日记能力——日记 CRUD、心情/天气、多图、仅本人可见

### Modified Capabilities
<!-- 无既有 spec 变更 -->

## Impact

- **新增模块**：`internal/diary/`（model.go / service.go / handler.go / handler_test.go）
- **修改**：`cmd/server/main.go`（注册 diary 路由 + AutoMigrate）
- **私密语义**：所有接口用 JWTMiddleware（非 OptionalJWT），service 层强制 author_id == currentUser
- **复用**：图片走既有 `/upload/image`；分页/DTO 模式照搬 posts
- **前端**：本 change 不含前端；前端在 diary-frontend change
- **不做**：日记不进公开信息流、无点赞/评论/收藏（私密）、无草稿态（日记直接保存）
- **mood/weather**：存字符串（如 mood="happy"/weather="sunny"），合法值前端约定，后端不强校验（灵活）
