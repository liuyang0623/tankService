## Context

go-service 现有 posts 模块是成熟的内容 CRUD 参考：Post 模型（Title/Content/Cover/Images 子表）、PostService（Create/FindAll/FindOne/Update/Remove/FindMyPosts，事务处理图片）、PostHandler（JWT 取 userID、参数解析、DTO 响应）、路由在 main.go authorized 组。已有 upload 图片能力、JWTMiddleware、response 统一包裹。

现状要点：
- posts 用 `authorized.Use(JWTMiddleware)` 组 + service 层 userID 校验归属（如 Update 校验 AuthorID == userID）
- PostImage 多图子表：foreignKey PostID + OnDelete CASCADE + SortOrder
- 分页 PaginationMeta + PaginatedResult DTO 模式
- AutoMigrate 在 main.go 列出所有模型

## Goals / Non-Goals

**Goals:**
- Diary 模型 + 多图子表
- 私密 CRUD（仅本人）
- mood/weather 字段
- 时间线列表（我的，倒序分页）

**Non-Goals:**
- 不进公开信息流
- 无点赞/评论/收藏（私密）
- 无草稿态（直接保存）
- mood/weather 不强校验合法值（前端约定）
- 不改前端

## Decisions

### D1. Diary 模型（照搬 Post 模式，去掉社交字段）
```go
type Diary struct {
    gorm.Model
    Title    string
    Content  string `gorm:"type:text"` // 富文本 HTML
    Cover    string // 封面图 URL（正文首图）
    Mood     string `gorm:"type:varchar(20)"`  // 心情，如 happy/sad
    Weather  string `gorm:"type:varchar(20)"`  // 天气，如 sunny/rainy
    AuthorID uint   `gorm:"index"`
    Images []DiaryImage `gorm:"foreignKey:DiaryID;constraint:OnDelete:CASCADE;"`
}
type DiaryImage struct {
    gorm.Model
    DiaryID   uint
    URL       string
    SortOrder int `gorm:"default:0"`
}
```
- 无 Status（无草稿）、无 ViewCount/LikeCount/CommentCount（私密无社交）、无 Topics/Category

### D2. Service 私密强制
- 所有查询 `Where("author_id = ?", userID)`
- FindOne/Update/Remove 先校验 Diary.AuthorID == userID，否则返回 ErrRecordNotFound（不泄露存在性）
- Create 事务：建 Diary → 建 DiaryImage（照搬 posts Create）

### D3. REST 接口（JWT，仅本人）
```
POST   /diaries        创建
GET    /diaries        我的列表（分页，created_at DESC）
GET    /diaries/:id    详情（校验归属）
PATCH  /diaries/:id    更新（校验归属）
DELETE /diaries/:id    删除（校验归属）
```
- 全部在 authorized 组（JWTMiddleware，非 OptionalJWT）

### D4. DTO
- `DiaryResponse`：id/title/content/cover/mood/weather/images/createdAt/updatedAt
- `DiaryListItem`：id/title/contentPreview/cover/mood/weather/createdAt（列表精简）
- `CreateDiaryInput`/`UpdateDiaryInput`：title/content/cover/mood/weather/images

## Risks / Open Questions

- mood/weather 空值允许（用户可不打卡）
- 富文本 content 与 posts 一样存 HTML，前端 RichEditor 复用
- 图片顺序：SortOrder 按前端传入顺序（照搬 posts extractImagesInOrder 语义）
