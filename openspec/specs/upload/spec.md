# upload Specification

## Purpose
TBD - created by archiving change go-service. Update Purpose after archive.
## Requirements
### Requirement: 图片上传至又拍云
系统 SHALL 允许已认证用户上传图片文件（jpeg/png/gif/webp），大小限制 10MB；使用 HMAC-SHA1 签名认证上传至又拍云；返回文件访问 URL。

#### Scenario: 上传图片成功
- **WHEN** POST `/api/v1/upload/image`，multipart/form-data，字段名 `file`，为有效图片
- **THEN** 文件上传至又拍云，返回 `{url, path, filename, mimetype, size}`

#### Scenario: 文件类型不允许
- **WHEN** POST `/api/v1/upload/image`，文件类型非图片
- **THEN** 返回 400，提示仅支持图片格式

#### Scenario: 文件过大
- **WHEN** POST `/api/v1/upload/image`，文件超过 10MB
- **THEN** 返回 400，提示文件大小超限

### Requirement: 通用文件上传至又拍云
系统 SHALL 允许已认证用户上传任意文件，大小限制 50MB；返回文件访问 URL。

#### Scenario: 上传文件成功
- **WHEN** POST `/api/v1/upload/file`，multipart/form-data，字段名 `file`
- **THEN** 文件上传至又拍云，返回 `{url, path, filename, mimetype, size}`

#### Scenario: 文件过大
- **WHEN** POST `/api/v1/upload/file`，文件超过 50MB
- **THEN** 返回 400，提示文件大小超限

### Requirement: 又拍云 HMAC-SHA1 签名
系统 SHALL 按又拍云 REST API 规范生成认证签名：对密码 MD5，用密码 MD5 对 `METHOD&/bucket/uri&date` 做 HMAC-SHA1，Base64 编码后拼接 `Authorization: UPYUN operator:signature`。

#### Scenario: 签名生成正确
- **WHEN** 调用签名生成函数，传入 method、uri、date
- **THEN** 返回符合又拍云规范的签名字符串

