# Proposal: 修复 profile 接口在用户不存在时错误返回 404

## 问题描述

线上 `GET /api/v1/users/profile` 返回 `{"data":null,"code":404,"message":"user not found"}`。

复现场景：本地开发环境登录后保留了本地签发的 JWT，切换到线上环境继续请求时，
JWT 签名校验通过（`JWTMiddleware` 放行），但 token 内的 `userID` 在线上数据库中
不存在对应用户，导致 `GetProfile` 命中 `gorm.ErrRecordNotFound` 分支，返回 404。

## 根因分析

`internal/users/handler.go` 的 `GetProfile` 与 `UpdateProfile` 在 service 返回
`gorm.ErrRecordNotFound` 时统一返回 `404 "user not found"`。

对于「凭 JWT 取当前登录用户」这类接口，token 有效但用户在库中不存在，本质是
**登录态失效**（例如账号已被删除、或 token 来自另一套环境/数据库），语义上应等同于
未授权，返回 `401` 触发前端重新登录，而不是返回 `404` 让前端误以为是资源缺失。

对比：`GET /users/:id`（按路径 ID 查任意用户）返回 404 是正确的——那是查询他人资源，
资源不存在就是 404。区别在于 profile 类接口的「主体是当前登录用户自身」。

## 修复目标

1. `GetProfile`：当当前登录用户在库中不存在时，返回 `401`（登录失效语义），
   而非 `404`。
2. `UpdateProfile`：同理，当前登录用户不存在时返回 `401`。
3. 不改动 `GetUser`（按路径 ID 查他人）的 404 行为——那是正确的资源不存在语义。
4. 补充/调整 handler 单测覆盖新的 401 行为。

## 影响范围

- 单模块 `internal/users`，改动集中在 `handler.go` 的两处错误分支 + 对应单测。
- 对外契约变化：profile 接口在「当前用户不存在」场景下状态码 `404 → 401`；
  前端据此走重新登录流程，符合预期修复目标。
