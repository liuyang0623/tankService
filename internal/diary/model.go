package diary

import (
	"time"

	"gorm.io/gorm"
)

// Diary represents a private diary entry.
// All diaries are private (only visible to the author).
type Diary struct {
	gorm.Model
	Title    string
	Content  string `gorm:"type:text"` // rich text HTML
	Cover    string // cover image URL
	Mood     string `gorm:"type:varchar(20)"`  // mood emoji key, e.g. "happy", "sad"
	Weather  string `gorm:"type:varchar(20)"`  // weather key, e.g. "sunny", "rainy"
	AuthorID uint   `gorm:"index"`
	NotebookID uint `gorm:"index"` // 归属日记本，0=无归属

	Images []DiaryImage `gorm:"foreignKey:DiaryID;constraint:OnDelete:CASCADE;"`
}

// TableName returns the table name for Diary.
func (Diary) TableName() string {
	return "diaries"
}

// DiaryImage represents an image attached to a diary entry.
type DiaryImage struct {
	gorm.Model
	DiaryID   uint
	URL       string
	SortOrder int `gorm:"default:0"`
}

// TableName returns the table name for DiaryImage.
func (DiaryImage) TableName() string {
	return "diary_images"
}

// --- DTOs ---

// DiaryResponse is the API response format for a single diary entry.
type DiaryResponse struct {
	ID        uint          `json:"id"`
	Title     string        `json:"title"`
	Content   string        `json:"content"`
	Cover     string        `json:"cover,omitempty"`
	Mood      string        `json:"mood,omitempty"`
	Weather   string        `json:"weather,omitempty"`
	NotebookID uint         `json:"notebookId"`
	Images    []ImageInfo   `json:"images,omitempty"`
	CreatedAt time.Time     `json:"createdAt"`
	UpdatedAt time.Time     `json:"updatedAt"`
}

// DiaryListItem is a simplified response for the diary timeline list.
type DiaryListItem struct {
	ID             uint      `json:"id"`
	Title          string    `json:"title"`
	ContentPreview string    `json:"contentPreview"`
	Cover          string    `json:"cover,omitempty"`
	Mood           string    `json:"mood,omitempty"`
	Weather        string    `json:"weather,omitempty"`
	NotebookID     uint      `json:"notebookId"`
	CreatedAt      time.Time `json:"createdAt"`
}

// ImageInfo is a simplified image representation in the API response.
type ImageInfo struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"order"`
}

// CreateDiaryInput holds the input for creating a diary entry.
type CreateDiaryInput struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Cover   string   `json:"cover"`
	Mood    string   `json:"mood"`
	Weather string   `json:"weather"`
	NotebookID uint  `json:"notebookId"`
	Images  []string `json:"images"`
}

// UpdateDiaryInput holds the input for updating a diary entry.
// All fields are optional (pointer or nil).
type UpdateDiaryInput struct {
	Title   *string   `json:"title,omitempty"`
	Content *string   `json:"content,omitempty"`
	Cover   *string   `json:"cover,omitempty"`
	Mood    *string   `json:"mood,omitempty"`
	Weather *string   `json:"weather,omitempty"`
	NotebookID *uint  `json:"notebookId,omitempty"`
	Images  *[]string `json:"images,omitempty"`
}

// PaginatedResult wraps a paginated list of diary items with meta info.
type PaginatedResult struct {
	Data []DiaryListItem `json:"data"`
	Meta PaginationMeta  `json:"meta"`
}

// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}
