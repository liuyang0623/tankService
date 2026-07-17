## Why

目前 go-service 缺少独立的开发环境配置文件，开发者无法直接获取本地开发配置。本次变更在 go-service 中新增对应的开发环境配置，简化本地开发启动流程。

## What Changes

- 新增 `go-service/.env.development` 文件，包含本地开发所需环境变量
- 更新 `go-service/.gitignore`，忽略 `.env.development`，避免敏感配置被提交到版本控制
- 不修改业务代码逻辑、数据库 schema 或 API 行为

## Capabilities

### New Capabilities

- `dev-env-config`: 开发环境配置管理，包含数据库、JWT、微信、又拍云等本地开发所需环境变量

### Modified Capabilities

无

## Impact

- 新增文件：`go-service/.env.development`
- 修改文件：`go-service/.gitignore`
- 不影响现有 API 行为或生产环境配置
