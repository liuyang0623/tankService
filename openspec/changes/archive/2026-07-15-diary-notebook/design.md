## Context

go-service 现有 diary 模块（Change diary-backend）：Diary 模型（Title/Content/Cover/Mood/Weather/AuthorID + Images 子表）、DiaryService（私密 CRUD）、DiaryHandler（JWT）。posts 模块提供成熟参照。本 change 补 notebook 并让 diary 归属 notebook。

## Goals / Non-Goals

**Goals:**
- Notebook 模型 + 私密 CRUD
- 默认本自动创建
- diary 归属 notebook + 按本过滤

**Non-Goals:**
- 不做日记本封面上传流程（前端复用 uploadApi，后端只存 URL）
- 不级联删日记（删本时日记 notebook_id 置 0）
- 不做日记本排序/置顶（第二期）

## Decisions

### D1. Notebook 模型（照搬 diary 私密模式）
```go
type Notebook struct {
    gorm.Model
    Name     string `gorm:"type:varchar(50)"`
    Color    string `gorm:"type:varchar(20)"`
    Cover    string
    AuthorID uint   `gorm:"index"`
}
```

### D2. 默认本
- `DefaultNotebookName = "默认"`
- `EnsureDefault(userID)`：count==0 时建默认本（color=#f0a868）
- FindMine 先 EnsureDefault 再查，保证用户永远至少一个本

### D3. Service 私密强制
- 所有查询 `Where("author_id = ?", userID)`
- Update/Remove 先校验归属，越权返回 ErrRecordNotFound

### D4. REST 接口（JWT，仅本人）
```
POST   /notebooks       创建
GET    /notebooks       我的列表（含 diaryCount，自动建默认本）
PATCH  /notebooks/:id   更新
DELETE /notebooks/:id   删除（本内日记 notebook_id 置 0）
```

### D5. diary 归属
- Diary 加 `NotebookID uint gorm:"index"`
- DiaryResponse/DiaryListItem/CreateDiaryInput 加 notebookId；UpdateDiaryInput 加 `*uint`
- FindMine 增加可选 notebookID 过滤（0=不过滤）
- handler FindMine 读 `c.Query("notebookId")`

## Risks / Open Questions

- diaryCount 用子查询实时算（notebook 数量少，性能可接受）
- 删本后孤儿日记（notebook_id=0）第一期归入"无归属"，前端默认本兜底展示
