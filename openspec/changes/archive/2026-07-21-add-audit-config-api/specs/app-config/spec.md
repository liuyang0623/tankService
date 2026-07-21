# app-config Specification

## ADDED Requirements

### Requirement: 全局应用配置查询
系统 SHALL 提供一个只读的 GET 接口，返回全局应用配置对象，供小程序前端在启动阶段拉取。该接口 MUST NOT 要求身份认证（在用户登录前即被调用），返回结构遵循项目统一响应契约：成功时 `code=200`、`message="success"`、配置数据置于 `data` 字段内，字段名采用项目统一的 camelCase json tag 风格。

#### Scenario: 拉取全局配置
- **WHEN** GET `/api/v1/app-config`（无需 JWT）
- **THEN** 返回 `{data: {auditMode: <bool>}, code: 200, message: "success"}`

#### Scenario: 免鉴权访问
- **WHEN** 未携带 Authorization 头调用该接口
- **THEN** 正常返回配置对象，不返回 401

### Requirement: 审核模式开关
配置对象 MUST 包含布尔字段 `auditMode`：`true` 表示当前处于审核模式（前端应隐藏/降级点赞、采纳、灵感互动等社交互动板块）；`false` 表示正常模式（全功能开放）。开关值 SHALL 来源于数据库单行配置表，支持运行时更新且不依赖前后端发版即时生效。

#### Scenario: 正常模式
- **WHEN** DB 中 `audit_mode` 为 false 时请求该接口
- **THEN** 返回 `data.auditMode = false`

#### Scenario: 审核模式
- **WHEN** 运营将 DB 中 `audit_mode` 更新为 true 后请求该接口
- **THEN** 返回 `data.auditMode = true`，无需重启服务或前端发版

### Requirement: 配置缺失时的种子默认
当数据库中尚无配置记录时，系统 SHALL 返回种子默认值 `auditMode = false`（正常模式），并 SHOULD 在首次启动时写入一行默认配置，保证接口始终可返回确定结果。

#### Scenario: 首次启动无记录
- **WHEN** 数据库 `app_config` 表为空时请求该接口
- **THEN** 返回 `data.auditMode = false`，不报错

### Requirement: 前端降级默认态约定
前端消费方在接口不可用、超时或响应解析失败时 MUST 采用兜底默认 `auditMode = true`（按审核模式处理，隐藏互动板块）。此约定与后端在线时的种子默认 `false` 不冲突：后端健康时返回真实值，后端不可达时前端自行假设审核模式以规避审核期风险。

#### Scenario: 接口超时或不可用
- **WHEN** 前端请求 `/api/v1/app-config` 超时、网络失败或响应无法解析
- **THEN** 前端按 `auditMode = true` 处理，隐藏/降级互动板块
