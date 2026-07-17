package follow

import "gorm.io/gorm"

// Follow represents a "user follows user" relationship.
// FollowerID is the user who follows; FollowingID is the user being followed.
type Follow struct {
	gorm.Model
	FollowerID  uint `gorm:"not null;index:idx_follow_pair,unique"`
	FollowingID uint `gorm:"not null;index:idx_follow_pair,unique"`
}

// TableName returns the database table name for Follow.
func (Follow) TableName() string {
	return "follows"
}

// FollowUserItem is the simplified user info returned in follower/following lists.
type FollowUserItem struct {
	ID          uint   `json:"id"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	IsFollowing bool   `json:"isFollowing"` // whether the current logged-in user follows this item
}

// PaginationMeta contains pagination metadata (same shape as other modules).
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// PaginatedUsers wraps a paginated list of follow user items.
type PaginatedUsers struct {
	Data []FollowUserItem `json:"data"`
	Meta PaginationMeta   `json:"meta"`
}
