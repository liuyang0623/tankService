# inspiration-sport

## ADDED Requirements

### Requirement: 用户可以创建运动目标
The system SHALL allow an authenticated user to create a personal sport goal with a name, optional type/icon, and an optional target day count.

#### Scenario: 成功创建目标
- **WHEN** 已认证用户以合法名称提交创建运动目标请求
- **THEN** 系统创建目标记录，关联所属用户 ID，初始连续天数与总打卡天数为 0，返回目标详情，`code=200`

#### Scenario: 名称为空拒绝
- **WHEN** 已认证用户提交空名称的创建请求
- **THEN** 系统返回 `400` 并提示名称必填

#### Scenario: 未认证拒绝
- **WHEN** 未携带有效 JWT 的请求调用创建接口
- **THEN** 系统返回 `401`

### Requirement: 用户可以查看自己的运动目标列表
The system SHALL return only the current user's own sport goals, ordered by creation time descending.

#### Scenario: 获取本人目标列表
- **WHEN** 已认证用户请求运动目标列表
- **THEN** 系统仅返回该用户创建的目标，每项含名称、目标天数、连续打卡天数、总打卡天数、今日是否已打卡

#### Scenario: 不返回他人目标
- **WHEN** 已认证用户请求运动目标列表
- **THEN** 结果中不包含其他用户创建的目标

### Requirement: 用户可以更新运动目标
The system SHALL allow the owner to update a goal's editable fields (name, type, target day count).

#### Scenario: 成功更新
- **WHEN** 目标所有者提交合法更新字段
- **THEN** 系统更新对应字段并返回最新目标详情

#### Scenario: 更新他人目标拒绝
- **WHEN** 用户尝试更新非本人创建的目标
- **THEN** 系统返回 `404`（视作不可见）

### Requirement: 用户可以按天打卡并累计连续天数
The system SHALL record a daily check-in for a goal owned by the user, computing consecutive-day streak and total check-in days. Multiple check-ins on the same calendar day MUST be idempotent (count once).

#### Scenario: 首次打卡
- **WHEN** 目标所有者当天第一次对目标打卡
- **THEN** 系统创建当日打卡记录，总打卡天数 +1，连续天数按规则更新，返回最新进度

#### Scenario: 同日重复打卡幂等
- **WHEN** 目标所有者在同一自然日内再次打卡
- **THEN** 系统不重复计数，总打卡天数与连续天数保持不变，返回当前进度

#### Scenario: 连续天数累加
- **WHEN** 目标所有者在前一日已打卡的基础上于次日打卡
- **THEN** 连续天数在原基础上 +1

#### Scenario: 中断后连续天数重置
- **WHEN** 目标所有者上次打卡距今超过一个自然日（漏打）后再次打卡
- **THEN** 连续天数重置为 1，总打卡天数照常 +1

#### Scenario: 打卡他人目标拒绝
- **WHEN** 用户尝试对非本人创建的目标打卡
- **THEN** 系统返回 `404`
