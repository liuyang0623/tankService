## Context

tankService 现有模块（article、diary、user 等）统一采用 handler / service / model 三层 + gin + gorm + `pkg/response` 封装，路由在 `cmd/server/main.go` 的 `authorized` 组注册，模型经 `database.AutoMigrate` 建表。本 change 新增 `internal/inspiration/` 模块，为「灵感」tab 的解惑问答与运动计划提供 API，完全复用现有约定，不引入新依赖。

已确认业务规则：
- 解惑 = 全站公开互助（列表返回全站问题，任意用户可回答任意问题）
- 运动计划 = 按天连续打卡（记录每日打卡，统计连续天数 + 总天数，同日幂等）

## Goals / Non-Goals

**Goals:**
- 提供解惑问答 CRUD 子集：提问、全站列表分页、详情（含回答）、回答
- 提供运动计划：目标 CRUD 子集 + 按天打卡与连续天数统计
- 与 diary 模块保持一致的分层、错误处理、响应封装、鉴权方式
- service 层核心逻辑（尤其连续打卡判定）有单元测试覆盖

**Non-Goals:**
- 问题/回答的点赞、评论、举报、删除、编辑（本期不做）
- 运动目标的社交分享、排行榜
- 打卡的补卡、跨时区精确处理（按服务器本地自然日）

## Decisions

### 模块结构
单一 `internal/inspiration` 包，内含四类模型与两组 service/handler，路由前缀分别为 `/questions` 和 `/sport-goals`。选择合并为一个包而非拆两个包：两块同属「灵感」tab、体量小，合并减少样板；若后续膨胀再拆分。

### 数据模型（GORM）
- `Question`：`gorm.Model` + `AuthorID(index)`、`Title`、`Content(text)`；`Answers []Answer` 关联；`AnswerCount` 通过查询计数不落库。
- `Answer`：`gorm.Model` + `QuestionID(index)`、`AuthorID(index)`、`Content(text)`。
- `SportGoal`：`gorm.Model` + `UserID(index)`、`Name`、`Type`、`Icon`、`TargetDays`、`Streak`（当前连续天数）、`TotalDays`（总打卡天数）、`LastCheckinDate`（date，判连续用）。
- `SportRecord`：`gorm.Model` + `GoalID(index)`、`UserID(index)`、`CheckinDate(date, index)`；`(goal_id, checkin_date)` 唯一，保证同日幂等。

将 `Streak`/`TotalDays`/`LastCheckinDate` 冗余在 `SportGoal` 上，避免每次列表都聚合 `SportRecord`；打卡时在事务内同步更新。

### 连续打卡判定算法（打卡事务内）
以服务器本地自然日 `today` 为基准：
1. 若 `(goal_id, today)` 记录已存在 → 幂等返回，进度不变。
2. 否则创建当日 `SportRecord`，`TotalDays += 1`：
   - `LastCheckinDate == today - 1`（昨天）→ `Streak += 1`（连续）
   - `LastCheckinDate == today` → 不应发生（步骤1已拦截）
   - 其他（含首次打卡、漏打）→ `Streak = 1`（重置）
3. `LastCheckinDate = today`，保存。
全程 `db.Transaction` 包裹，唯一索引兜底并发重复打卡。

### 路由（authorized 组，均需 JWT）
```
POST   /questions              提问
GET    /questions              全站问题列表（分页）
GET    /questions/:id          问题详情 + 回答列表
POST   /questions/:id/answers  回答问题
GET    /sport-goals            我的目标列表
POST   /sport-goals            创建目标
PATCH  /sport-goals/:id        更新目标
POST   /sport-goals/:id/checkin 打卡
```

### 复用现有约定
- `getUserID(c)` 从 gin context 取 JWT 注入的 `userID`（与 diary 同名 helper，包内私有各自定义）
- 分页复用 `parsePagination` 同款逻辑（page 默认 1、limit 默认 10）
- 错误：`response.BadRequest / Unauthorized / InternalError / Error(404)`；`gorm.ErrRecordNotFound` → 404
- 所有权隔离：sport-goal 的 update/checkin 用 `WHERE user_id = ?` 过滤，查不到即 404

## Risks / Trade-offs

- [冗余字段与 SportRecord 不一致] → 所有写入都在同一事务内更新 goal 冗余字段；不提供绕过 service 的写路径。
- [跨时区/服务器时区打卡边界] → 本期按服务器本地自然日，接受边界误差；文档标注为 Non-Goal。
- [并发同日重复打卡] → `(goal_id, checkin_date)` 唯一索引兜底，事务内先查后插。
- [问题列表 N+1 计数回答数] → 列表用一次 `GROUP BY question_id` 聚合回答数，避免逐条查询。

## Migration Plan

- 部署：`AutoMigrate` 自动新建 4 张表，无历史数据迁移，无破坏性变更。
- 回滚：新增表与路由互不影响既有功能，回滚即移除路由注册与模型注册；表可保留或手动 drop。

## Open Questions

- 无阻塞项。问题/回答是否需要软删除、运动目标完成后的归档态，留待前端联调后按需再开 change。
