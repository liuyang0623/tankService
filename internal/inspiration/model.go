package inspiration

import (
	"time"

	"gorm.io/gorm"
)

// --- Models: 解惑问答 ---

// Question 是用户提出的一个问题（全站公开互助）。
type Question struct {
	gorm.Model
	AuthorID uint   `gorm:"index"`
	Title    string `gorm:"type:varchar(200)"`
	Content  string `gorm:"type:text"`

	Answers []Answer `gorm:"foreignKey:QuestionID;constraint:OnDelete:CASCADE;"`
}

// TableName 返回 Question 的表名。
func (Question) TableName() string {
	return "questions"
}

// Answer 是某个用户对某个问题的回答。任意用户可回答任意问题。
type Answer struct {
	gorm.Model
	QuestionID uint   `gorm:"index"`
	AuthorID   uint   `gorm:"index"`
	Content    string `gorm:"type:text"`
}

// TableName 返回 Answer 的表名。
func (Answer) TableName() string {
	return "answers"
}

// --- Models: 运动计划 ---

// SportGoal 是用户创建的运动目标，冗余记录连续/累计打卡进度以便列表直接读取。
type SportGoal struct {
	gorm.Model
	UserID     uint   `gorm:"index"`
	Name       string `gorm:"type:varchar(100)"`
	Type       string `gorm:"type:varchar(30)"` // 运动类型 key，如 running/yoga
	Icon       string `gorm:"type:varchar(30)"` // 图标 key
	TargetDays int    `gorm:"default:0"`        // 目标打卡天数，0=不限
	Streak     int    `gorm:"default:0"`        // 当前连续打卡天数
	TotalDays  int    `gorm:"default:0"`        // 累计打卡天数

	// LastCheckinDate 记录最近一次打卡的自然日（00:00），用于判定连续。
	LastCheckinDate *time.Time `gorm:"type:date"`
}

// TableName 返回 SportGoal 的表名。
func (SportGoal) TableName() string {
	return "sport_goals"
}

// SportRecord 是一次每日打卡记录。(goal_id, checkin_date) 唯一，保证同日幂等。
type SportRecord struct {
	gorm.Model
	GoalID      uint      `gorm:"index:idx_goal_date,unique"`
	UserID      uint      `gorm:"index"`
	CheckinDate time.Time `gorm:"type:date;index:idx_goal_date,unique"`
}

// TableName 返回 SportRecord 的表名。
func (SportRecord) TableName() string {
	return "sport_records"
}

// --- DTOs: 解惑问答 ---

// CreateQuestionInput 是创建问题的入参。
type CreateQuestionInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// CreateAnswerInput 是回答问题的入参。
type CreateAnswerInput struct {
	Content string `json:"content"`
}

// AnswerResponse 是回答的 API 响应格式。
type AnswerResponse struct {
	ID        uint      `json:"id"`
	AuthorID  uint      `json:"authorId"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"createdAt"`
}

// QuestionResponse 是问题详情的 API 响应格式（含回答列表）。
type QuestionResponse struct {
	ID          uint             `json:"id"`
	AuthorID    uint             `json:"authorId"`
	Title       string           `json:"title"`
	Content     string           `json:"content"`
	AnswerCount int              `json:"answerCount"`
	Answers     []AnswerResponse `json:"answers"`
	CreatedAt   time.Time        `json:"createdAt"`
	UpdatedAt   time.Time        `json:"updatedAt"`
}

// QuestionListItem 是问题列表的简化响应。
type QuestionListItem struct {
	ID          uint      `json:"id"`
	AuthorID    uint      `json:"authorId"`
	Title       string    `json:"title"`
	AnswerCount int       `json:"answerCount"`
	CreatedAt   time.Time `json:"createdAt"`
}

// --- DTOs: 运动计划 ---

// CreateSportGoalInput 是创建运动目标的入参。
type CreateSportGoalInput struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Icon       string `json:"icon"`
	TargetDays int    `json:"targetDays"`
}

// UpdateSportGoalInput 是更新运动目标的入参，字段均可选。
type UpdateSportGoalInput struct {
	Name       *string `json:"name,omitempty"`
	Type       *string `json:"type,omitempty"`
	Icon       *string `json:"icon,omitempty"`
	TargetDays *int    `json:"targetDays,omitempty"`
}

// SportGoalResponse 是运动目标的 API 响应格式。
type SportGoalResponse struct {
	ID             uint       `json:"id"`
	Name           string     `json:"name"`
	Type           string     `json:"type,omitempty"`
	Icon           string     `json:"icon,omitempty"`
	TargetDays     int        `json:"targetDays"`
	Streak         int        `json:"streak"`
	TotalDays      int        `json:"totalDays"`
	CheckedInToday bool       `json:"checkedInToday"`
	LastCheckinAt  *time.Time `json:"lastCheckinAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

// CheckinResponse 是打卡后的进度响应。
type CheckinResponse struct {
	GoalID         uint `json:"goalId"`
	Streak         int  `json:"streak"`
	TotalDays      int  `json:"totalDays"`
	CheckedInToday bool `json:"checkedInToday"`
	Awarded        bool `json:"awarded"` // 本次是否新增了打卡（false=同日重复）
}

// --- 分页 ---

// PaginatedResult 包装分页问题列表与元信息。
type PaginatedResult struct {
	Data []QuestionListItem `json:"data"`
	Meta PaginationMeta     `json:"meta"`
}

// PaginationMeta 是分页元信息。
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}
