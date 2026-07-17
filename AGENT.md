# AGENT.md

给 AI 协作者的**开发前置条件与铁律**。动手前先读这里；技术栈、脚本、目录、接口见 [README.md](./README.md)，本文件不重复，只讲"不看就会踩坑或违规"的部分。

本仓库是 **tankService**（微信小程序后端，Go+Gin+GORM+MySQL）。前端仓为 tankingMiniprogram。

---

## 0. 最高优先级：comet 工作流阶段门

本仓库用 **comet + OpenSpec** 管理需求到归档的全流程。**不要裸开发**——成体量的新特性/改动必须走 comet，不能直接 `openspec new change` 跳过阶段门，也不能直接写源码。

存在活跃 change 时（`openspec/changes/<name>/.comet.yaml` 存在），**每次动手前必须先读 `.comet.yaml` 的 `phase` 字段**，只做当前阶段允许的操作。

| 阶段 | 允许 | 禁止 |
|------|------|------|
| `open` | 建 proposal/design/tasks | 写源码 |
| `design` | brainstorming、写 Design Doc | 写源码 |
| `build` | 写源码、测试、执行计划 | 跳过用户确认点 |
| `verify` | 验证、分支处理 | 跳过失败处理 |
| `archive` | 确认归档、跑归档脚本 | 写源码 |

**硬性规则**（详见 `.claude/rules/comet-phase-guard.md`）：

- **comet 脚本在 `.claude/skills/comet/scripts/`**（本仓库路径；注意与前端仓的 `.comet/skills/skills/comet/scripts/` 不同）。
- **阶段推进只能经 guard/transition 脚本**，禁止手工编辑 `.comet.yaml` 或跳阶段。
- 阶段退出跑 `node .claude/skills/comet/scripts/comet-guard.mjs <name> <phase> --apply`，必须看到 `ALL CHECKS PASSED`。
- 归档跑 `node .claude/skills/comet/scripts/comet-archive.mjs <name>`（需先 guard 过 verify→archive）。
- 也可用 comet CLI：`comet status` 查活跃 change，`comet init` 重装基建。
- verify 阶段的 **branch handling** 是用户决策点，不能自动填。
- **判断有无活跃 change**：`ls openspec/changes/ | grep -v archive`。为空则无活跃 change——归档后的零散 bugfix/小改直接提交即可，不必凭空新建 change。

> 只有成体量的新特性才走完整 open→archive 流程；归档后的小修不必开新 change。启动新特性用 `/comet "想法"`（完整）、`/comet-hotfix`（快修）、`/comet-tweak`（小改）。

---

## 1. 模块结构与约定（照现有模式加新模块）

每个业务模块在 `internal/<name>/`，固定三件套 + 测试：
- `model.go` — GORM 模型（`gorm.Model` 内嵌）+ 响应 DTO + `TableName()`
- `service.go` — 业务逻辑，`New<Name>Service(db)` 构造，方法签名 `(ctx, ...)`，错误用 `fmt.Errorf("...: %w", err)` 包装
- `handler.go` — HTTP handler，`service` 接口抽象便于 mock 测试；`getUserID(c)` 从 JWT 取用户；统一用 `pkg/response`（`Success`/`Unauthorized`/`BadRequest`/`Error`）
- `handler_test.go` — mock service 测 handler（鉴权/路由/响应）

**装配在 `cmd/server/main.go`**：`setupRouter` 里 `New<Name>Service(db)` + 路由挂 `authorized`（JWT）或 optional/public 组；`main()` 的 `AutoMigrate(...)` 列表加新模型。

**跨模块依赖**：单向 + 接口注入避免包循环。范例：`follow` 需要写通知，通过 `follow` 内定义的 `notifier` 接口注入 `notification` service（`followSvc.SetNotifier(notificationSvc)`），notification 不反向依赖 follow。

**响应契约**：成功 `{ code: 200, data, message }`；分页 `{ data, meta:{total,page,limit,totalPages} }`。

---

## 2. 数据库坑（真机连 MySQL 才暴露）

- **MySQL 保留字**：列名避开 `read`/`order`/`group` 等保留字。GORM 生成的 `WHERE read=?` 会报 1064 语法错，且 mock 测试碰不到（handler 层不连真实 DB）。用 `column:` tag 映射成非保留字（如 `Read bool` → `column:is_read`）。这类问题**只有真机联调暴露**，写涉及新列的查询时先自检列名。
- 建表靠 `main.go` 的 `AutoMigrate`；改了模型字段/列名后**必须重启服务**才会建列。AutoMigrate 只增不删（改列名会残留旧列，无害）。
- 测试用 mock service，不接真实 DB（项目未引 sqlite）。service 层的 SQL 逻辑靠 `go build` + 真机联调保证，不要假设 mock 测试能防住 SQL 问题。

---

## 3. 提交前必过的验证

```bash
go build ./...    # 编译，必须通过
go test ./...     # 全量测试，保持全绿
```

- 涉及新模块/新接口，补 handler 层单测（参考各模块 `handler_test.go`）。
- DB 相关改动（新列、新查询、migration）如实说明"需真机连 MySQL 验证"，不要声称已验证。

---

## 4. 环境与安全

- 敏感配置在 `.env`（已 gitignore），模板见 `.env.example`。**绝不提交 `.env`、密钥、`WECHAT_SECRET`、`UPYUN_PASSWORD` 等**。
- go module 名是 `go-service`（历史保留，与仓库名 tankService 不一致属正常，Go 允许）。import 路径仍是 `go-service/...`，不要改。

---

## 5. Git 约定

- 提交信息 `<type>(scope): 描述`，type ∈ feat/fix/refactor/docs/test/chore/perf/ci。
- **提交/推送需用户明确要求**，不要自动 push。

---

## 6. 文档同步：重大功能更新必须更新 README

**重大功能更新（新模块、新接口、鉴权变更、目录结构调整、技术栈/依赖变化等），必须同步更新 [README.md](./README.md)**，随代码改动一起提交。

- 判断标准：改动影响到 README 已描述的内容（技术栈、目录结构、可用命令、环境变量、主要接口表），就一并更新对应段落。
- 新增业务模块 → README 的目录结构 + 主要接口表补上。
- 纯 bugfix、内部重构等不改变对外描述的改动，不必动 README。
- 若变更涉及 AI 铁律（新平台约束、新数据库坑、新跨模块模式），回头也更新本文件（AGENT.md）。
