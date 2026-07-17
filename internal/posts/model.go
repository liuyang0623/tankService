package posts

import (
	"time"

	"gorm.io/gorm"
)

// PostStatus represents the publication status of a post.
type PostStatus string

const (
	// PostStatusDraft is the default status for new posts.
	PostStatusDraft PostStatus = "DRAFT"
	// PostStatusPublished means the post is publicly visible.
	PostStatusPublished PostStatus = "PUBLISHED"
)

// Fixed post categories. Empty category means uncategorized.
const (
	CategoryStory  = "story"
	CategoryDaily  = "daily"
	CategoryTech   = "tech"
	CategoryFood   = "food"
	CategoryTravel = "travel"
)

// validCategories is the set of allowed category values.
var validCategories = map[string]bool{
	CategoryStory:  true,
	CategoryDaily:  true,
	CategoryTech:   true,
	CategoryFood:   true,
	CategoryTravel: true,
}

// IsValidCategory reports whether c is a known category. Empty string is valid
// (uncategorized); any other unknown value is invalid.
func IsValidCategory(c string) bool {
	if c == "" {
		return true
	}
	return validCategories[c]
}

// CategoryInfo is the API representation of a category (value + display label).
type CategoryInfo struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// FixedCategories is the ordered list returned by GET /categories.
var FixedCategories = []CategoryInfo{
	{Value: CategoryStory, Label: "故事"},
	{Value: CategoryDaily, Label: "日常"},
	{Value: CategoryTech, Label: "技术"},
	{Value: CategoryFood, Label: "美食"},
	{Value: CategoryTravel, Label: "旅游"},
}

// PostAuthor is a lightweight author representation loaded via GORM Preload.
// Maps to the "users" table but only selects id, nickname, avatar.
type PostAuthor struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// TableName maps PostAuthor to the users table.
func (PostAuthor) TableName() string {
	return "users"
}

// Post represents a blog/article post in the system.
type Post struct {
	gorm.Model
	Title        string
	Content      string `gorm:"type:text"`
	Cover        string // cover image URL
	Status       PostStatus `gorm:"default:DRAFT"`
	Category     string     `gorm:"type:varchar(20);index"` // fixed category, empty = uncategorized
	AuthorID     uint
	ViewCount    int `gorm:"default:0"`
	LikeCount    int `gorm:"default:0"`
	CommentCount int `gorm:"default:0"`
	PublishedAt  *time.Time

	Author PostAuthor  `gorm:"foreignKey:AuthorID;references:ID"`
	Images []PostImage `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;"`
	Topics []Topic     `gorm:"many2many:post_topics;"`
}

// TableName returns the table name for Post.
func (Post) TableName() string {
	return "posts"
}

// PostImage represents an image attached to a post.
type PostImage struct {
	gorm.Model
	PostID    uint
	URL       string
	SortOrder int `gorm:"default:0"`
}

// TableName returns the table name for PostImage.
func (PostImage) TableName() string {
	return "post_images"
}

// Topic represents a category/tag for posts.
type Topic struct {
	gorm.Model
	Name string `gorm:"type:varchar(191);uniqueIndex;not null"`
}

// TableName returns the table name for Topic.
func (Topic) TableName() string {
	return "topics"
}
