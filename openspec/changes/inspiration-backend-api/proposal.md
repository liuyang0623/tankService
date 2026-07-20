## Why

新增「灵感」tab 需要两个后端能力支撑：**解惑**（用户互助问答）和**运动计划**（运动目标与打卡）。现有 tankService 仅覆盖文章、日记、用户等模块，缺少这两类互动数据的持久化与 API，前端板块无法落地。

## What Changes

- 新增 `internal/inspiration/` 模块，遵循现有 diary 模块的 handler / service / model 三层模式。
- **解惑问答**（全站公开互助）：用户可提问、浏览**全站**问题列表（分页）、查看问题详情（含回答列表）、对**任意**问题作答。
- **运动计划**（按天连续打卡）：用户可创建运动目标、查看自己的目标列表、更新目标、按日打卡；系统记录每日打卡并统计**连续天数**与**总打卡天数**（同一天重复打卡幂等）。
- 在 `cmd/server/main.go` 的 `authorized` 路由组注册新路由（全部需 JWT 认证）。
- 在 `AutoMigrate` 注册新增数据模型（questions / answers / sport_goals / sport_records）。
- 所有响应沿用统一封装：成功 `code=200`，分页返回 `{data, meta}`。

## Capabilities

### New Capabilities
- `inspiration-qa`: 解惑问答能力——全站公开互助，问题的创建、全站列表分页、详情、任意用户回答。
- `inspiration-sport`: 运动计划能力——运动目标的创建、列表、更新，以及按天连续打卡记录（连续天数 + 总天数统计）。

### Modified Capabilities
<!-- 无现有能力的需求变更 -->

## Impact

- **新增代码**：`internal/inspiration/{model,service,handler,service_test}.go`
- **修改代码**：`cmd/server/main.go`（路由注册 + AutoMigrate）
- **数据库**：新增 4 张表 `questions`、`answers`、`sport_goals`、`sport_records`（GORM AutoMigrate 自动建表，无破坏性变更）
- **API**：`/api/v1` 下新增 questions、sport-goals 相关端点，均需 Bearer JWT
- **依赖**：无新增第三方依赖，复用 gin / gorm / pkg/response
