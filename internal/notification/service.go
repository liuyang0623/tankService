package notification

import (
	"context"
	"fmt"
	"math"

	"go-service/internal/users"

	"gorm.io/gorm"
)

// NotificationService handles system notification storage and queries.
type NotificationService struct {
	db *gorm.DB
}

// NewNotificationService creates a new NotificationService.
func NewNotificationService(db *gorm.DB) *NotificationService {
	return &NotificationService{db: db}
}

// Notification type constants (extensible).
const (
	TypeFollow = "follow"
)

// CreateFollow writes a "follow" notification for recipient userID triggered by actorID.
// Callers should treat failures as non-fatal (log only), so as not to block the
// primary action (following).
func (s *NotificationService) CreateFollow(ctx context.Context, userID, actorID uint) error {
	if userID == actorID {
		return nil // never notify self
	}
	n := &Notification{UserID: userID, Type: TypeFollow, ActorID: actorID}
	if err := s.db.WithContext(ctx).Create(n).Error; err != nil {
		return fmt.Errorf("create follow notification: %w", err)
	}
	return nil
}

// List returns the user's notifications ordered by newest first (paginated),
// with each actor's display info attached.
func (s *NotificationService) List(ctx context.Context, userID uint, page, limit int) (*PaginatedNotifications, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.WithContext(ctx).Model(&Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count notifications: %w", err)
	}

	var notifs []Notification
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifs).Error; err != nil {
		return nil, fmt.Errorf("find notifications: %w", err)
	}

	items, err := s.attachActors(ctx, notifs)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	return &PaginatedNotifications{
		Data: items,
		Meta: PaginationMeta{Total: total, Page: page, Limit: limit, TotalPages: totalPages},
	}, nil
}

// MarkAllRead marks all of the user's unread notifications as read (idempotent).
func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint) error {
	if err := s.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("mark all read: %w", err)
	}
	return nil
}

// UnreadSummary returns the unread count and the latest notification (for the
// aggregated entry). Latest is nil when the user has no notifications.
func (s *NotificationService) UnreadSummary(ctx context.Context, userID uint) (*UnreadSummary, error) {
	var unread int64
	if err := s.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&unread).Error; err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}

	var latest Notification
	err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&latest).Error
	if err == gorm.ErrRecordNotFound {
		return &UnreadSummary{UnreadCount: unread, Latest: nil}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find latest: %w", err)
	}

	items, err := s.attachActors(ctx, []Notification{latest})
	if err != nil {
		return nil, err
	}
	return &UnreadSummary{UnreadCount: unread, Latest: &items[0]}, nil
}

// attachActors batch-loads actor display info and builds NotificationItem slices.
func (s *NotificationService) attachActors(ctx context.Context, notifs []Notification) ([]NotificationItem, error) {
	items := make([]NotificationItem, len(notifs))
	if len(notifs) == 0 {
		return items, nil
	}

	actorIDs := make([]uint, 0, len(notifs))
	seen := map[uint]bool{}
	for _, n := range notifs {
		if !seen[n.ActorID] {
			seen[n.ActorID] = true
			actorIDs = append(actorIDs, n.ActorID)
		}
	}

	var us []users.User
	if err := s.db.WithContext(ctx).Where("id IN ?", actorIDs).Find(&us).Error; err != nil {
		return nil, fmt.Errorf("load actors: %w", err)
	}
	actorMap := make(map[uint]ActorInfo, len(us))
	for _, u := range us {
		actorMap[u.ID] = ActorInfo{ID: u.ID, Nickname: u.Nickname, Avatar: u.Avatar}
	}

	for i, n := range notifs {
		items[i] = NotificationItem{
			ID:        n.ID,
			Type:      n.Type,
			Actor:     actorMap[n.ActorID],
			TargetID:  n.TargetID,
			Read:      n.Read,
			CreatedAt: n.CreatedAt,
		}
	}
	return items, nil
}
