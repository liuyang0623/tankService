package message

import (
	"time"

	"go-service/internal/users"

	"gorm.io/gorm"
)

// Conversation represents a private chat between two users.
// UserAID < UserBID is enforced at creation time so (user_a_id, user_b_id) is unique.
type Conversation struct {
	gorm.Model
	UserAID     uint      `gorm:"not null;index;column:user_a_id"`
	UserBID     uint      `gorm:"not null;index;column:user_b_id"`
	LastMessage string    `gorm:"type:text"`
	LastTime    time.Time
	UserAUnread int       `gorm:"default:0"`
	UserBUnread int       `gorm:"default:0"`
}

// TableName returns the database table name for Conversation.
func (Conversation) TableName() string {
	return "conversations"
}

// Message represents a single message within a conversation.
type Message struct {
	gorm.Model
	ConversationID uint   `gorm:"not null;index"`
	SenderID       uint   `gorm:"not null"`
	Type           string `gorm:"type:varchar(20);not null;default:'text'"`
	Content        string `gorm:"type:text;not null"`
	Read           bool   `gorm:"default:false"`
}

// TableName returns the database table name for Message.
func (Message) TableName() string {
	return "messages"
}

// ConversationItem is the response shape for a conversation in the list.
type ConversationItem struct {
	ID           uint          `json:"id"`
	OtherUser    OtherUserInfo `json:"otherUser"`
	LastMessage  string        `json:"lastMessage"`
	LastTime     time.Time     `json:"lastTime"`
	UnreadCount  int           `json:"unreadCount"`
}

// OtherUserInfo is a minimal user representation for conversation items.
type OtherUserInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

// MessageItem is the response shape for a single message.
type MessageItem struct {
	ID             uint      `json:"id"`
	ConversationID uint      `json:"conversationId"`
	SenderID       uint      `json:"senderId"`
	Type           string    `json:"type"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"createdAt"`
}

// PaginatedConversations wraps a paginated conversation list.
type PaginatedConversations struct {
	Data []ConversationItem `json:"data"`
	Meta PaginationMeta     `json:"meta"`
}

// PaginatedMessages wraps a paginated message list.
type PaginatedMessages struct {
	Data []MessageItem `json:"data"`
	Meta PaginationMeta     `json:"meta"`
}

// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// SendMessageRequest is the request body for POST /messages.
type SendMessageRequest struct {
	ToUserID uint   `json:"toUserId" binding:"required"`
	Type     string `json:"type" binding:"omitempty,oneof=text image"`
	Content  string `json:"content" binding:"required"`
}

// ensureConversation creates or returns an existing conversation between two users.
// It sorts IDs so UserAID < UserBID for a unique pair.
func ensureConversation(db *gorm.DB, userA, userB uint) (*Conversation, error) {
	aid, bid := userA, userB
	if aid > bid {
		aid, bid = bid, aid
	}

	var conv Conversation
	err := db.Where("user_a_id = ? AND user_b_id = ?", aid, bid).First(&conv).Error
	if err == nil {
		return &conv, nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	conv = Conversation{
		UserAID:  aid,
		UserBID:  bid,
		LastTime: time.Now(),
	}
	if err := db.Create(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// toConversationItem builds a ConversationItem from a Conversation and the other user's info.
func (c *Conversation) toConversationItem(other *users.User, currentUserID uint) ConversationItem {
	unread := c.UserAUnread
	if currentUserID == c.UserBID {
		unread = c.UserBUnread
	}
	return ConversationItem{
		ID: c.ID,
		OtherUser: OtherUserInfo{
			ID:       other.ID,
			Nickname: other.Nickname,
			Avatar:   other.Avatar,
		},
		LastMessage: c.LastMessage,
		LastTime:    c.LastTime,
		UnreadCount: unread,
	}
}

// toMessageItem builds a MessageItem from a Message.
func (m *Message) toMessageItem() MessageItem {
	return MessageItem{
		ID:             m.ID,
		ConversationID: m.ConversationID,
		SenderID:       m.SenderID,
		Type:           m.Type,
		Content:        m.Content,
		CreatedAt:      m.CreatedAt,
	}
}
