# inspiration-qa

## ADDED Requirements

### Requirement: 用户可以提问
The system SHALL allow an authenticated user to create a question with a title and optional body content.

#### Scenario: 成功提问
- **WHEN** 已认证用户以合法标题提交提问请求
- **THEN** 系统创建问题记录，关联提问者用户 ID，并返回问题详情，`code=200`

#### Scenario: 标题为空拒绝
- **WHEN** 已认证用户提交空标题的提问请求
- **THEN** 系统返回 `400` 并提示标题必填，不创建记录

#### Scenario: 未认证拒绝
- **WHEN** 未携带有效 JWT 的请求调用提问接口
- **THEN** 系统返回 `401`

### Requirement: 用户可以浏览全站问题列表
The system SHALL return a paginated list of ALL users' questions, ordered by creation time descending, regardless of who created them.

#### Scenario: 分页获取全站问题
- **WHEN** 已认证用户请求问题列表，指定 `page` 与 `limit`
- **THEN** 系统返回 `{data, meta}`，`data` 为按创建时间倒序的全站问题列表，`meta` 含 `total/page/limit/totalPages`

#### Scenario: 列表项包含回答数
- **WHEN** 已认证用户请求问题列表
- **THEN** 每个列表项包含问题标题、提问者信息、回答数量与创建时间

#### Scenario: 分页参数缺省
- **WHEN** 请求未提供 `page` 或 `limit`
- **THEN** 系统使用默认 `page=1`、`limit=10`

### Requirement: 用户可以查看问题详情
The system SHALL return a single question with its full body and all associated answers ordered by creation time ascending.

#### Scenario: 查看存在的问题
- **WHEN** 已认证用户请求某个存在的问题 ID
- **THEN** 系统返回该问题详情，含标题、正文、提问者、回答列表，`code=200`

#### Scenario: 查看不存在的问题
- **WHEN** 已认证用户请求不存在的问题 ID
- **THEN** 系统返回 `404`

### Requirement: 用户可以回答任意问题
The system SHALL allow any authenticated user to submit an answer to any question (public mutual help), including questions created by other users.

#### Scenario: 成功回答
- **WHEN** 已认证用户对某存在问题提交非空回答内容
- **THEN** 系统创建回答记录，关联回答者用户 ID 与问题 ID，并返回回答详情，`code=200`

#### Scenario: 回答内容为空拒绝
- **WHEN** 已认证用户提交空内容的回答
- **THEN** 系统返回 `400` 并提示内容必填

#### Scenario: 回答不存在的问题
- **WHEN** 已认证用户对不存在的问题 ID 提交回答
- **THEN** 系统返回 `404`
