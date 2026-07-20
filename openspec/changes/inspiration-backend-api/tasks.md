# Tasks: inspiration-backend-api

## 1. 数据模型

- [x] 1.1 创建 `internal/inspiration/model.go`，定义 `Question`、`Answer` 模型（含 TableName、索引）
- [x] 1.2 在 model.go 定义 `SportGoal`、`SportRecord` 模型，`(goal_id, checkin_date)` 唯一索引
- [x] 1.3 定义请求/响应 DTO：`CreateQuestionInput`、`QuestionResponse`、`QuestionListItem`、`CreateAnswerInput`、`AnswerResponse`、`PaginatedResult`、`PaginationMeta`
- [x] 1.4 定义运动 DTO：`CreateSportGoalInput`、`UpdateSportGoalInput`、`SportGoalResponse`、`CheckinResponse`

## 2. Service 层 — 解惑问答

- [x] 2.1 创建 `internal/inspiration/service.go` 的 `QAService`（含 `NewQAService`）
- [x] 2.2 实现 `CreateQuestion`（校验标题非空，关联 AuthorID）
- [x] 2.3 实现 `ListQuestions`（全站分页，倒序，聚合回答数，返回 PaginatedResult）
- [x] 2.4 实现 `GetQuestion`（含回答列表，正序；不存在返回 gorm.ErrRecordNotFound）
- [x] 2.5 实现 `CreateAnswer`（校验内容非空，校验问题存在，关联 QuestionID/AuthorID）

## 3. Service 层 — 运动计划

- [x] 3.1 实现 `SportService`（含 `NewSportService`）
- [x] 3.2 实现 `CreateGoal`（校验名称非空，初始化 Streak/TotalDays=0）
- [x] 3.3 实现 `ListGoals`（仅本人，倒序，含今日是否已打卡标记）
- [x] 3.4 实现 `UpdateGoal`（所有权校验，查不到即 ErrRecordNotFound）
- [x] 3.5 实现 `Checkin`（事务内：同日幂等、连续天数判定、更新冗余字段）

## 4. Handler 层

- [x] 4.1 创建 `internal/inspiration/handler.go`，实现 `getUserID` helper 与 `parsePagination`
- [x] 4.2 实现 `QAHandler`：Create / List / FindOne / CreateAnswer（错误映射 400/401/404/500）
- [x] 4.3 实现 `SportHandler`：ListGoals / CreateGoal / UpdateGoal / Checkin

## 5. 路由与迁移接入

- [x] 5.1 在 `cmd/server/main.go` 的 `authorized` 组注册 questions / sport-goals 路由
- [x] 5.2 在 `AutoMigrate` 注册 `Question`、`Answer`、`SportGoal`、`SportRecord`

## 6. 测试与验证

- [x] 6.1 创建 `internal/inspiration/service_test.go`，覆盖 QA：提问/列表/详情/回答，及空校验、404
- [x] 6.2 覆盖运动打卡关键路径：首次打卡、同日幂等、连续+1、中断重置、他人目标 404
- [x] 6.3 运行 `go test ./internal/inspiration/...` 全绿
- [x] 6.4 运行 `go build ./...` 编译通过
