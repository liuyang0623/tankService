## Why

小程序端联调发现三个交互相关的后端缺陷，导致点赞/收藏、评论、详情用户态无法正常工作。均已通过复现与代码定位确证根因，需在后端修复。

## What Changes

- **点赞/收藏重复操作 500（BREAKING bug）**：`Like`/`Favorite` 使用 `gorm.Model`（软删除）+ `(post_id, user_id)` 唯一索引。toggle 时软删除记录仍占用唯一索引，第三次操作 `Create` 触发唯一键冲突报 500。改为硬删除（`Unscoped().Delete`）或复用软删除记录，使 toggle 可无限次正常切换。
- **评论接口返回 GORM 原始模型**：`GetPostComments`/`CreateComment` 直接返回 `Comment` 模型，字段为 PascalCase（`ID`/`Content`/`UserID`），且不含作者信息。改为返回带 `author`、小写 json 字段的 Comment DTO，与其它接口（PostResponse）风格一致，前端才能正确回显评论与作者。
- **详情接口拿不到当前用户态**：`GET /posts/:id` 为纯公开路由，无鉴权中间件，`optionalUserID` 永远为空，`isLiked`/`isFavorited` 恒为 false。新增「可选鉴权中间件」（有 token 则解析注入 userID，无 token 放行），挂到公开的详情/列表等路由，使登录用户能拿到真实互动态。

不改变对外 API 的 URL 与整体契约，只修正响应字段与行为正确性。

## Capabilities

### New Capabilities

- `interaction-fixes`: 点赞/收藏 toggle 可无限次正常切换（修复软删除唯一索引冲突）；评论列表/创建返回带作者信息的 DTO；公开详情/列表路由支持可选鉴权，登录用户获得 `isLiked`/`isFavorited` 真实态

### Modified Capabilities

无（go-service 主 spec 尚未建立对应能力，本次以新增 spec 记录修复后的目标行为）

## Impact

- `internal/interactions/service.go`：LikePost/FavoritePost 删除逻辑；GetPostComments/CreateComment 返回 DTO
- `internal/interactions/handler.go` / `model.go`：Comment DTO 定义
- `pkg/middleware/jwt.go`：新增 `OptionalJWTMiddleware`
- `cmd/server/main.go`：公开路由挂可选鉴权中间件
- 数据库：无 schema 变更（唯一索引保留，改删除策略）
- 前端 `tankingMiniprogram` 依赖本修复：详情改 authRequest、评论回显（本 change 完成后进行）
