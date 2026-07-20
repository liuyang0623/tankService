## Context

现有关注订阅消息链路：`follow` 触发关注事件 → `subscribepush.Pusher.pushFollowSync` 查配额/openid → 调 `wechat.Client.SendSubscribeMessage(ctx, openid, tplID, data)` 下发。

当前 `SendSubscribeMessage` 构造的请求 body 只有 `touser`、`template_id`、`data` 三个字段，未携带微信 `subscribe/send` 接口支持的 `page` 参数，因此用户点击订阅消息后只能打开小程序默认首页。

前端小程序（`tankingMiniprogram`）已存在系统通知页 `pages/notifications/index`（`app.config.ts` 中注册，标题「系统通知」）。该页进入时调用 `GET /notifications` 拉取列表并整体标记已读，自给自足，不依赖任何「单条通知详情」接口。

## Goals / Non-Goals

**Goals:**
- 让关注订阅消息携带 `page`，用户点击后直达系统通知页 `pages/notifications/index`。
- `SendSubscribeMessage` 具备透传任意 `page` 的能力，为后续其他类型订阅消息复用。
- 固化 `page` 路径契约，确保与前端真实页面一致。

**Non-Goals:**
- 不新增单条通知详情接口（前端列表页自给自足）。
- 不改通知模型、通知类型、配额/授权逻辑。
- 不动前端代码。

## Decisions

1. **`SendSubscribeMessage` 签名扩展方式**：增加 `page string` 参数（放在 `data` 之后）。仅当 `page != ""` 时才写入 body 的 `page` 字段。空字符串不写入，保证对其他调用点（若有）和微信接口的向后兼容——微信 `page` 为可选字段，缺省即不跳转。

2. **page 路径取 `pages/notifications/index` 而非 `pages/notification/detail`**：经核对前端 `app.config.ts` 与 `pages/notifications/index.tsx`，小程序实际注册并可打开的通知页是复数形式的列表页 `pages/notifications/index`；`pages/notification/detail` 页面不存在，深链过去会打不开。因此契约定为前端真实存在的路径。

3. **深链取数不需要新接口**：系统通知页深链进入时执行 `loadList()` + `markRead()`，只依赖现有 `GET /notifications`。用户从微信空降到该页也能正常渲染，无需按 id 取单条，故后端零取数改动。

4. **page 值放在 Pusher 内**：`pushFollowSync` 调用 sender 时传入常量 `pages/notifications/index`，与模板字段拼装逻辑同处一地，便于后续按通知类型扩展不同跳转页。

## Risks / Trade-offs

- **路径漂移风险**：若前端后续重命名通知页路由，后端常量需同步更新。已在契约中显式记录路径，降低失联风险。
- **微信 page 校验**：微信要求 `page` 必须是小程序已发布的合法页面路径，否则下发可能报错。`pages/notifications/index` 为已注册页面，风险低；page 为空时不写入字段，不会因此导致下发失败。
- **sender 接口变更影响面**：`SendSubscribeMessage` 属于 `wechat.Client`，同时被 `subscribepush` 的 `sender` 接口引用。扩展参数需同步更新接口定义与测试桩，改动集中在 `internal/wechat` 与 `internal/subscribepush` 两处。
