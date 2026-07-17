package notification

import (
	"time"

	"gorm.io/gorm"
)

// Notification represents a system notification delivered to a user.
// It is one-directional (system/actor -> recipient), distinct from private
// messages. Type is extensible (follow now; like/comment reserved).
type Notification struct {
	gorm.Model
	UserID   uint   `gorm:"not null;index:idx_notif_user_read;index:idx_notif_user_created"` // recipient
	Type     string `gorm:"type:varchar(20);not null"`                                       // follow / like / comment ...
	ActorID  uint   `gorm:"not null"`                                                        // who triggered it
	TargetID uint   `gorm:"default:0"`                                                       // optional: post id etc. (reserved)
	Read     bool   `gorm:"column:is_read;default:false;index:idx_notif_user_read"` // "read" is a MySQL reserved word → map to is_read
}

// TableName returns the database table name for Notification.
func (Notification) TableName() string {
	return "notifications"
}

// ActorInfo is the minimal trigger-user representation embedded in a notification item.
type ActorInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// NotificationItem is the response shape for a single notification.
type NotificationItem struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Actor     ActorInfo `json:"actor"`
	TargetID  uint      `json:"targetId,omitempty"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"createdAt"`
}

// PaginationMeta contains pagination metadata (same shape as other modules).
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// PaginatedNotifications wraps a paginated notification list.
type PaginatedNotifications struct {
	Data []NotificationItem `json:"data"`
	Meta PaginationMeta     `json:"meta"`
}

// UnreadSummary is returned by the unread-count endpoint to drive the
// aggregated system-notification entry on the messages page.
type UnreadSummary struct {
	UnreadCount int64             `json:"unreadCount"`
	Latest      *NotificationItem `json:"latest"` // most recent notification, or null if none
}
