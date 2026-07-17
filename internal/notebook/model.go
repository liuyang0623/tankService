package notebook

import (
	"time"

	"gorm.io/gorm"
)

// DefaultNotebookName 是新用户自动创建的默认日记本名称。
const DefaultNotebookName = "默认"

// Notebook 日记本，私密（仅本人）。
type Notebook struct {
	gorm.Model
	Name     string `gorm:"type:varchar(50)"`
	Color    string `gorm:"type:varchar(20)"` // 封面/兜底色，如 #f0a868
	Cover    string // 日记本封面图 URL（可选）
	AuthorID uint   `gorm:"index"`
}

// TableName 返回表名。
func (Notebook) TableName() string { return "notebooks" }

// NotebookResponse 是日记本 API 响应。
type NotebookResponse struct {
	ID         uint      `json:"id"`
	Name       string    `json:"name"`
	Color      string    `json:"color"`
	Cover      string    `json:"cover,omitempty"`
	DiaryCount int64     `json:"diaryCount"`
	CreatedAt  time.Time `json:"createdAt"`
}

// CreateNotebookInput 创建入参。
type CreateNotebookInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Cover string `json:"cover"`
}

// UpdateNotebookInput 更新入参（可选字段）。
type UpdateNotebookInput struct {
	Name  *string `json:"name,omitempty"`
	Color *string `json:"color,omitempty"`
	Cover *string `json:"cover,omitempty"`
}
