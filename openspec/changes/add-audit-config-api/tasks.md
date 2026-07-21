# add-audit-config-api 实施任务

## 1. 后端 appconfig 能力包

- [x] 1.1 新建 `internal/appconfig/model.go`：`AppConfig` gorm 模型（固定单行，主键 + `AuditMode bool` 列，DB 列名 `audit_mode`）；`AppConfigResponse` DTO（`AuditMode bool \`json:"auditMode"\``）
- [x] 1.2 `internal/appconfig/service.go`：`GetConfig(ctx)` 读取单行配置；无记录时返回种子默认 `auditMode=false`（不报错）
- [x] 1.3 `internal/appconfig/handler.go`：`GET /app-config` handler，调用 service 后 `response.Success(c, resp)`
- [x] 1.4 单测：有记录返回真实值、无记录返回 false 默认、DTO json tag 为 `auditMode`（camelCase）

## 2. 迁移与装配

- [x] 2.1 `app_config` 表加入 `AutoMigrate`；首次启动 seed 一行 `audit_mode=false`（幂等，仅当表空时插入）
- [x] 2.2 `cmd/server/main.go`：构造 appconfig service/handler，在公开路由组注册 `v1.GET("/app-config", ...)`（免鉴权，与 `/auth/wechat/login` 同级，不进 `authorized` 组）
- [x] 2.3 单测：免鉴权可访问（无 JWT 返回 200 而非 401）；响应体为 `{data:{auditMode:...}, code:200, message:"success"}`

## 3. 后端验证

- [x] 3.1 `go build ./...` 通过
- [x] 3.2 `go test ./...` 全绿

## 4. 前端对接（另仓 tankingMiniprogram，本仓仅记录约定）

- [ ] 4.1 启动流程拉取 `GET /api/v1/app-config`，按 `data.auditMode` 控制互动板块渲染
- [ ] 4.2 实现降级兜底：请求超时/失败/解析失败时按 `auditMode=true`（审核模式）处理

## 5. 收尾

- [x] 5.1 按文档同步规则更新后端 README（新增 app-config 接口表、目录、切换说明：DB update `app_config.audit_mode`）
- [x] 5.2 记录运营切换步骤（送审前置 true、通过后置 false）到 README 或运维说明
