## Why

diary-frontend（前端日记功能，Moo 风格）需要「日记本」概念：日记归属日记本，用户可切换/新建/管理日记本。现有 diary 表（Change diary-backend 交付）只有单篇日记，无分组。本 change 补齐后端日记本能力。

## What Changes

- **新增 notebook 模块**（`internal/notebook/`）：Notebook 模型 + service + handler
- **notebook 表**：name/color/cover/authorID，私密（仅本人）
- **notebook CRUD 接口**：POST/GET /notebooks、PATCH/DELETE /notebooks/:id（均 JWT）
- **默认本策略**：用户首次 GET /notebooks 时若无任何日记本，自动创建"默认"本
- **diary 加 notebook_id**：Diary 模型加 NotebookID 字段；GET /diaries 支持 ?notebookId= 过滤；create/update 接受 notebookId
- **删除策略**：删日记本时，本内日记 notebook_id 置 0（不级联删日记）

## Capabilities

### New Capabilities
- `diary-notebook`: 后端日记本能力——notebook CRUD、diary 归属日记本、默认本、按本过滤

### Modified Capabilities
<!-- diary 能力的接口有微调（加 notebookId），但作为 diary-notebook 的一部分，不单列 -->

## Impact

- **新增模块**：`internal/notebook/`（model/service/handler + test）
- **修改**：`internal/diary/`（model 加 NotebookID，service 加过滤，handler 读 query）、`cmd/server/main.go`（notebook 路由 + AutoMigrate）
- **私密语义**：notebook 所有接口 JWTMiddleware + service 层归属校验，越权 404
- **前端依赖**：diary-frontend change 消费本 change 的 /notebooks 接口
