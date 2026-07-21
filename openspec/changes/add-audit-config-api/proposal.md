## Why

小程序上架到微信/各应用市场送审时，审核期间需要临时隐藏或降级部分 UGC/社交互动功能（点赞、采纳、灵感互动板块等），避免因未成年人保护、内容合规等审核红线被打回；审核通过后再恢复。

目前前端没有统一的开关来源，只能靠发版切换——审核期临时改一个开关也要走完整发版流程，无法即时收口。需要后端提供一个全局配置下发接口，让前端启动时拉取、按开关即时决定渲染，不再依赖发版。

## What Changes

- 新增 **app-config（全局应用配置）能力**：后端提供一个**只读**的全局配置查询接口，前端启动时拉取。
- 核心开关 `auditMode`（布尔）：`true` = 审核模式，前端隐藏/降级互动板块；`false` = 正常模式，全功能开放。
- **配置存 DB，运行时可切**：新增单行配置表 `app_config`，改值只需更新一行（DB update / 运营后台），**不依赖前后端发版即时生效**。DB 种子初始值为 `false`（正常模式）；送审前运营翻成 `true`，审核通过翻回 `false`。
- **免鉴权**：该接口在用户登录前、小程序启动阶段即被调用，不携带 JWT，直接挂在公开路由组（与 `POST /auth/wechat/login` 同级）。
- **响应契约复用既有 `pkg/response.Success`**：成功 `{data:{...}, code:200, message:"success"}`，字段用项目统一的 **camelCase** json tag 风格（`auditMode`，而非 snake_case）。

### 降级语义（前端约定，需写入前端对接说明）

- **前端兜底默认 = `true`（审核模式）**：接口不可用/超时/解析失败时，前端 MUST 按审核模式处理（隐藏互动板块）——审核期更安全，宁可多藏不可漏出。
- **后端 DB 种子默认 = `false`（正常模式）**：后端正常在线时返回数据库真实值，平时为 `false`。
- 二者不矛盾：后端健康 → 报真实状态（平时 false）；后端不可达 → 前端才假设 true。

## Capabilities

### New Capabilities
- `app-config`: 全局应用配置下发能力——只读 GET 接口返回全局配置对象（含 `auditMode` 审核开关），配置存 DB 单行表、运行时可切、免鉴权、复用统一响应契约。

### Modified Capabilities
<!-- 不改任何现有 capability 的对外契约。互动相关能力（interactions / inspiration）的渲染开关由前端消费本接口决定，属前端行为，不改后端 spec 级要求。 -->

## Impact

- **后端（tankService）**：
  - 新增 `internal/appconfig/`：`model.go`（`AppConfig` 表 + `AppConfigResponse` DTO）、`service.go`（读取单行配置，缺失时返回种子默认）、`handler.go`（`GET /api/v1/app-config`）。
  - `app_config` 表加入 `AutoMigrate`；首次启动 seed 一行 `audit_mode=false`。
  - `cmd/server/main.go` 在公开路由组注册 `GET /app-config`（免鉴权）。
- **前端（小程序，另仓）**：启动流程拉取 `GET /api/v1/app-config`，按 `data.auditMode` 决定互动板块渲染；实现兜底默认 `true`（超时/失败按审核模式）。
- **配置**：无新增环境变量；开关值改由 DB 承载。运营切换方式：DB update `app_config` 或后续后台（本期最小实现只保证读接口 + DB 可改）。
- **依赖**：无新增第三方依赖；复用 gorm、`pkg/response`、gin 公开路由模式。
- **约束**：本期为**只读下发 + DB 可改值**，不含运营后台 UI、不含写接口/鉴权后台；单行表按固定主键读取。
