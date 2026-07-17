package users

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// userDB abstracts database operations needed by UserService.
// This interface enables injecting a fake implementation in tests.
type userDB interface {
	First(dest interface{}, id uint) error
	Updates(dest interface{}, values map[string]interface{}) error
}

// gormUserDB adapts *gorm.DB to the userDB interface.
type gormUserDB struct {
	db *gorm.DB
}

func (g *gormUserDB) First(dest interface{}, id uint) error {
	return g.db.First(dest, id).Error
}

func (g *gormUserDB) Updates(dest interface{}, values map[string]interface{}) error {
	return g.db.Model(dest).Updates(values).Error
}

// allowedUpdateFields is the whitelist of user profile fields that can be updated.
var allowedUpdateFields = map[string]bool{
	"nickname": true,
	"avatar":   true,
	"bio":      true,
	"gender":   true,
	"phone":    true,
}

// UserService handles user profile queries and updates.
type UserService struct {
	db userDB
}

// NewUserService creates a UserService backed by a real *gorm.DB.
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: &gormUserDB{db: db},
	}
}

// newUserServiceFromDB is used internally and in tests to inject a userDB directly.
func newUserServiceFromDB(db userDB) *UserService {
	return &UserService{db: db}
}

// GetProfile retrieves the user profile by userID.
// Returns gorm.ErrRecordNotFound if the user does not exist.
func (s *UserService) GetProfile(ctx context.Context, userID uint) (*User, error) {
	var user User
	if err := s.db.First(&user, userID); err != nil {
		return nil, err
	}
	return &user, nil
}

// FindOne retrieves the user (public info) by userID.
// Returns gorm.ErrRecordNotFound if the user does not exist.
// FindOne is the public-facing lookup; future authorization differences
// from GetProfile can be added here without changing callers.
func (s *UserService) FindOne(ctx context.Context, userID uint) (*User, error) {
	var user User
	if err := s.db.First(&user, userID); err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates allowed profile fields for a user and returns the updated record.
// Only the fields in allowedUpdateFields (nickname, avatar, bio, gender) are accepted;
// all other keys in updates are silently dropped.
// Returns gorm.ErrRecordNotFound if the user does not exist.
func (s *UserService) UpdateProfile(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
	// First verify the user exists.
	var user User
	if err := s.db.First(&user, userID); err != nil {
		return nil, err
	}

	// Filter updates to whitelist only.
	filtered := make(map[string]interface{}, len(updates))
	for k, v := range updates {
		if allowedUpdateFields[k] {
			filtered[k] = v
		}
	}

	// Only call DB if there is something to update.
	if len(filtered) > 0 {
		if err := s.db.Updates(&user, filtered); err != nil {
			return nil, fmt.Errorf("update profile: %w", err)
		}
	}

	return &user, nil
}
