package message

import (
	"context"
	"fmt"
	"math"

	"go-service/internal/users"

	"gorm.io/gorm"
)

// MessageService handles business logic for conversations and messages.
type MessageService struct {
	db *gorm.DB
}

// NewMessageService creates a new MessageService.
func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

// CreateMessage creates a new message in the conversation between sender and receiver.
// It auto-creates the conversation on first message.
func (s *MessageService) CreateMessage(ctx context.Context, senderID, toUserID uint, msgType, content string) (*Message, error) {
	if senderID == toUserID {
		return nil, fmt.Errorf("cannot send message to yourself")
	}

	// Ensure receiver exists
	var userCount int64
	if err := s.db.Model(&users.User{}).Where("id = ?", toUserID).Count(&userCount).Error; err != nil {
		return nil, fmt.Errorf("check user: %w", err)
	}
	if userCount == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	// Find or create conversation
	conv, err := ensureConversation(s.db, senderID, toUserID)
	if err != nil {
		return nil, fmt.Errorf("ensure conversation: %w", err)
	}

	// Create the message
	msg := Message{
		ConversationID: conv.ID,
		SenderID:       senderID,
		Type:           msgType,
		Content:        content,
		Read:           false,
	}
	if msg.Type == "" {
		msg.Type = "text"
	}
	if err := s.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Update conversation metadata. For image messages, store a placeholder
	// so the conversation list shows "[图片]" instead of the raw URL.
	lastMessage := content
	if msg.Type == "image" {
		lastMessage = "[图片]"
	}
	updates := map[string]interface{}{
		"last_message": lastMessage,
		"last_time":    msg.CreatedAt,
	}
	if senderID == conv.UserAID {
		updates["user_b_unread"] = gorm.Expr("user_b_unread + 1")
	} else {
		updates["user_a_unread"] = gorm.Expr("user_a_unread + 1")
	}
	if err := s.db.WithContext(ctx).Model(conv).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update conversation: %w", err)
	}

	return &msg, nil
}

// FindConversationByUsers returns the conversation id between two users, or 0 if none exists.
func (s *MessageService) FindConversationByUsers(ctx context.Context, userID, otherUserID uint) (uint, error) {
	aid, bid := userID, otherUserID
	if aid > bid {
		aid, bid = bid, aid
	}
	var conv Conversation
	err := s.db.WithContext(ctx).
		Where("user_a_id = ? AND user_b_id = ?", aid, bid).
		First(&conv).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("find conversation: %w", err)
	}
	return conv.ID, nil
}
func (s *MessageService) GetConversations(ctx context.Context, userID uint, page, limit int) (*PaginatedConversations, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Conversation{}).
		Where("user_a_id = ? OR user_b_id = ?", userID, userID).
		Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count conversations: %w", err)
	}

	var convs []Conversation
	if err := s.db.Where("user_a_id = ? OR user_b_id = ?", userID, userID).
		Order("last_time DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&convs).Error; err != nil {
		return nil, fmt.Errorf("find conversations: %w", err)
	}

	// Collect other user IDs and load user info
	otherIDs := make([]uint, 0, len(convs))
	for _, c := range convs {
		otherID := c.UserBID
		if userID == c.UserBID {
			otherID = c.UserAID
		}
		otherIDs = append(otherIDs, otherID)
	}

	var others []users.User
	if len(otherIDs) > 0 {
		if err := s.db.Where("id IN ?", otherIDs).Find(&others).Error; err != nil {
			return nil, fmt.Errorf("load other users: %w", err)
		}
	}
	otherMap := make(map[uint]users.User, len(others))
	for _, u := range others {
		otherMap[u.ID] = u
	}

	data := make([]ConversationItem, len(convs))
	for i, c := range convs {
		otherID := c.UserBID
		if userID == c.UserBID {
			otherID = c.UserAID
		}
		other := otherMap[otherID]
		data[i] = c.toConversationItem(&other, userID)
	}

	return &PaginatedConversations{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// GetMessages returns paginated messages for a conversation, verifying the user is a participant.
func (s *MessageService) GetMessages(ctx context.Context, conversationID, userID uint, page, limit int) (*PaginatedMessages, error) {
	// Verify user is a participant
	var conv Conversation
	if err := s.db.First(&conv, conversationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, fmt.Errorf("find conversation: %w", err)
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return nil, fmt.Errorf("not a participant of this conversation")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Message{}).
		Where("conversation_id = ?", conversationID).
		Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count messages: %w", err)
	}

	var msgs []Message
	if err := s.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC, id DESC").
		Limit(limit).
		Offset(offset).
		Find(&msgs).Error; err != nil {
		return nil, fmt.Errorf("find messages: %w", err)
	}

	// Reverse so oldest-first within the page
	data := make([]MessageItem, len(msgs))
	for i, m := range msgs {
		data[len(msgs)-1-i] = m.toMessageItem()
	}

	return &PaginatedMessages{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// MarkRead marks all unread messages in a conversation as read for the given user.
func (s *MessageService) MarkRead(ctx context.Context, conversationID, userID uint) error {
	var conv Conversation
	if err := s.db.First(&conv, conversationID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return err
		}
		return fmt.Errorf("find conversation: %w", err)
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return fmt.Errorf("not a participant of this conversation")
	}

	// Determine the other user's ID
	otherID := conv.UserBID
	if userID == conv.UserBID {
		otherID = conv.UserAID
	}

	// Mark unread messages where this user is the receiver as read
	if err := s.db.WithContext(ctx).Model(&Message{}).
		Where("conversation_id = ? AND sender_id = ? AND `read` = ?", conversationID, otherID, false).
		Update("read", true).Error; err != nil {
		return fmt.Errorf("mark messages read: %w", err)
	}

	// Reset unread count
	updateField := "user_a_unread"
	if userID == conv.UserBID {
		updateField = "user_b_unread"
	}
	if err := s.db.WithContext(ctx).Model(&conv).Update(updateField, 0).Error; err != nil {
		return fmt.Errorf("reset unread count: %w", err)
	}

	return nil
}
