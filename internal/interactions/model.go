package interactions

import "gorm.io/gorm"

// Like represents a user liking a post.
type Like struct {
	gorm.Model
	PostID uint `gorm:"not null;index:idx_like_post_user,unique"`
	UserID uint `gorm:"not null;index:idx_like_post_user,unique"`
}

// TableName returns the table name for Like.
func (Like) TableName() string {
	return "likes"
}

// Favorite represents a user favoriting a post.
type Favorite struct {
	gorm.Model
	PostID uint `gorm:"not null;index:idx_fav_post_user,unique"`
	UserID uint `gorm:"not null;index:idx_fav_post_user,unique"`
}

// TableName returns the table name for Favorite.
func (Favorite) TableName() string {
	return "favorites"
}

// Comment represents a user comment on a post.
type Comment struct {
	gorm.Model
	PostID   uint    `gorm:"not null;index"`
	UserID   uint    `gorm:"not null;index"`
	Content  string  `gorm:"not null"`
	ParentID *uint   `gorm:"index"` // nullable, for replies
	Replies  []Comment `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;"`
}

// TableName returns the table name for Comment.
func (Comment) TableName() string {
	return "comments"
}
