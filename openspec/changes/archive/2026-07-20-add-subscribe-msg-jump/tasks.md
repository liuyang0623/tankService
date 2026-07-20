# Tasks: 新增关注订阅消息跳转能力（仅后端）

## 1. 扩展微信订阅消息发送支持 page

- [x] 1.1 修改 `internal/wechat/client.go` 的 `SendSubscribeMessage`：新增 `page string` 参数（置于 `data` 之后），仅当 `page != ""` 时写入请求 body 的 `page` 字段
- [x] 1.2 更新 `internal/subscribepush/pusher.go` 中 `sender` 接口定义，使其方法签名与新的 `SendSubscribeMessage` 一致

## 2. Pusher 下发关注通知时携带跳转页

- [x] 2.1 修改 `internal/subscribepush/pusher.go` 的 `pushFollowSync`：调用 sender 时传入跳转页常量 `pages/notifications/index`
- [x] 2.2 将跳转页路径定义为包内常量，便于后续按通知类型扩展

## 3. 测试

- [x] 3.1 更新 `internal/wechat/client_test.go`：断言 `page` 非空时请求 body 含 `page` 字段、为空时不含
- [x] 3.2 更新 `internal/subscribepush/pusher_test.go`：断言关注推送时传入的 page 为 `pages/notifications/index`

## 4. 验证

- [x] 4.1 运行 `go build ./...` 与 `go test ./internal/wechat/... ./internal/subscribepush/...`，确认全绿

<!-- code review (standard): Ready to proceed. 无 Critical/Important。
     2 条 Minor 接受不修（既有代码风格，非本次引入）：
     M1 测试中 r.Body.Read 未检查返回值（小体积 JSON 实践稳定）；
     M2 WithPage 断言用 strings.Contains 而非 JSON 解析（与既有测试风格一致）。 -->
