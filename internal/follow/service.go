package follow

import (
	"context"
	"fmt"
	"log"
	"math"

	"go-service/internal/users"

	"gorm.io/gorm"
)

// notifier abstracts writing a follow notification. Defined here (consumer side)
// so follow depends on notification only via interface — no import cycle.
type notifier interface {
	CreateFollow(ctx context.Context, userID, actorID uint) error
}

// subscribePusher abstracts pushing a WeChat subscribe message when someone is
// followed. Fire-and-forget: the implementation handles quota/openid lookup,
// async goroutine, timeout, and errors internally (logs only).
type subscribePusher interface {
	PushFollow(ctx context.Context, targetID, actorID uint)
}

// FollowService handles user follow relationships.
type FollowService struct {
	db       *gorm.DB
	notifier notifier
	pusher   subscribePusher
}

// NewFollowService creates a new FollowService.
func NewFollowService(db *gorm.DB) *FollowService {
	return &FollowService{db: db}
}

// SetNotifier injects the notification writer used to notify a user when
// someone follows them. Optional — if unset, follow proceeds without notifying.
func (s *FollowService) SetNotifier(n notifier) {
	s.notifier = n
}

// SetSubscribePusher injects the WeChat subscribe-message pusher. Optional —
// if unset, follow proceeds without pushing.
func (s *FollowService) SetSubscribePusher(p subscribePusher) {
	s.pusher = p
}

// ToggleFollow toggles the follow relationship from followerID to targetID.
// Returns true if now following, false if unfollowed.
func (s *FollowService) ToggleFollow(ctx context.Context, followerID, targetID uint) (bool, error) {
	if followerID == targetID {
		return false, fmt.Errorf("cannot follow yourself")
	}

	// Ensure target user exists
	var userCount int64
	if err := s.db.Model(&users.User{}).Where("id = ?", targetID).Count(&userCount).Error; err != nil {
		return false, fmt.Errorf("check user: %w", err)
	}
	if userCount == 0 {
		return false, gorm.ErrRecordNotFound
	}

	// Look up existing relation
	var existing Follow
	err := s.db.Where("follower_id = ? AND following_id = ?", followerID, targetID).First(&existing).Error
	if err == nil {
		// Already following -> unfollow (hard delete to avoid soft-delete + unique index conflict)
		if err := s.db.Unscoped().Delete(&existing).Error; err != nil {
			return false, fmt.Errorf("unfollow: %w", err)
		}
		return false, nil
	}
	if err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("query follow: %w", err)
	}

	// Not following -> create
	if err := s.db.Create(&Follow{FollowerID: followerID, FollowingID: targetID}).Error; err != nil {
		return false, fmt.Errorf("follow: %w", err)
	}

	// Notify the followed user. Non-fatal: a notification failure must not
	// roll back or fail the follow itself.
	if s.notifier != nil {
		if err := s.notifier.CreateFollow(ctx, targetID, followerID); err != nil {
			log.Printf("follow notification failed (target=%d actor=%d): %v", targetID, followerID, err)
		}
	}

	// Push a WeChat subscribe message (async, fire-and-forget). Failures never
	// affect the follow or the in-app notification.
	if s.pusher != nil {
		s.pusher.PushFollow(ctx, targetID, followerID)
	}
	return true, nil
}

// CountFollowers returns how many users follow the given user.
func (s *FollowService) CountFollowers(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := s.db.Model(&Follow{}).Where("following_id = ?", userID).Count(&n).Error
	return n, err
}

// CountFollowing returns how many users the given user follows.
func (s *FollowService) CountFollowing(ctx context.Context, userID uint) (int64, error) {
	var n int64
	err := s.db.Model(&Follow{}).Where("follower_id = ?", userID).Count(&n).Error
	return n, err
}

// IsFollowing reports whether currentUserID follows targetID.
func (s *FollowService) IsFollowing(ctx context.Context, currentUserID, targetID uint) (bool, error) {
	if currentUserID == 0 {
		return false, nil
	}
	var n int64
	err := s.db.Model(&Follow{}).
		Where("follower_id = ? AND following_id = ?", currentUserID, targetID).
		Count(&n).Error
	return n > 0, err
}

// ListFollowers returns users who follow userID (paginated), with isFollowing
// computed for currentUserID.
func (s *FollowService) ListFollowers(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error) {
	// followers = users whose id is in (select follower_id from follows where following_id = userID)
	sub := s.db.Model(&Follow{}).Select("follower_id").Where("following_id = ?", userID)
	return s.listUsersBySubquery(ctx, sub, currentUserID, page, limit)
}

// ListFollowing returns users that userID follows (paginated), with isFollowing
// computed for currentUserID.
func (s *FollowService) ListFollowing(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error) {
	// following = users whose id is in (select following_id from follows where follower_id = userID)
	sub := s.db.Model(&Follow{}).Select("following_id").Where("follower_id = ?", userID)
	return s.listUsersBySubquery(ctx, sub, currentUserID, page, limit)
}

// listUsersBySubquery paginates users whose id is in the given subquery,
// ordered by user id desc (most recent first), and computes isFollowing in batch.
func (s *FollowService) listUsersBySubquery(ctx context.Context, sub *gorm.DB, currentUserID uint, page, limit int) (*PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&users.User{}).Where("id IN (?)", sub).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var us []users.User
	if err := s.db.Where("id IN (?)", sub).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&us).Error; err != nil {
		return nil, fmt.Errorf("find users: %w", err)
	}

	// Batch compute isFollowing: which of these users does currentUserID follow?
	followedSet := map[uint]bool{}
	if currentUserID != 0 && len(us) > 0 {
		ids := make([]uint, len(us))
		for i, u := range us {
			ids[i] = u.ID
		}
		var followed []Follow
		if err := s.db.Where("follower_id = ? AND following_id IN ?", currentUserID, ids).
			Find(&followed).Error; err != nil {
			return nil, fmt.Errorf("batch isFollowing: %w", err)
		}
		for _, f := range followed {
			followedSet[f.FollowingID] = true
		}
	}

	data := make([]FollowUserItem, len(us))
	for i, u := range us {
		data[i] = FollowUserItem{
			ID:          u.ID,
			Nickname:    u.Nickname,
			Avatar:      u.Avatar,
			Bio:         u.Bio,
			IsFollowing: followedSet[u.ID],
		}
	}

	return &PaginatedUsers{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}
