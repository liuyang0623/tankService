package users

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

// --- fake DB that implements userDB ---

type fakeUserDB struct {
	// for First (by id lookup)
	firstUser  *User
	firstError error
	// for Save/Updates
	saveUser  *User
	saveError error
	// record calls
	lastFirstID    uint
	lastUpdatesMap map[string]interface{}
	// for IncrColumn
	incrError      error
	lastIncrUserID uint
	lastIncrColumn string
}

func (f *fakeUserDB) First(dest interface{}, id uint) error {
	f.lastFirstID = id
	if f.firstError != nil {
		return f.firstError
	}
	if f.firstUser != nil {
		u := dest.(*User)
		*u = *f.firstUser
	}
	return nil
}

func (f *fakeUserDB) Updates(dest interface{}, values map[string]interface{}) error {
	f.lastUpdatesMap = values
	if f.saveError != nil {
		return f.saveError
	}
	if f.saveUser != nil {
		u := dest.(*User)
		*u = *f.saveUser
	}
	return nil
}

func (f *fakeUserDB) IncrColumn(model interface{}, userID uint, column string) error {
	f.lastIncrUserID = userID
	f.lastIncrColumn = column
	return f.incrError
}

func TestIncrSubscribeFollowQuota_Success(t *testing.T) {
	fdb := &fakeUserDB{}
	svc := newUserServiceFromDB(fdb)
	if err := svc.IncrSubscribeFollowQuota(context.Background(), 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fdb.lastIncrUserID != 42 {
		t.Errorf("expected incr for user 42, got %d", fdb.lastIncrUserID)
	}
	if fdb.lastIncrColumn != "subscribe_follow_quota" {
		t.Errorf("expected column subscribe_follow_quota, got %q", fdb.lastIncrColumn)
	}
}

// --- GetProfile tests ---

func TestGetProfile_Success(t *testing.T) {
	fdb := &fakeUserDB{
		firstUser: &User{
			Model:    gorm.Model{ID: 1},
			Openid:   "wx-openid",
			Nickname: "Alice",
			Avatar:   "https://example.com/avatar.png",
			Bio:      "hello",
			Gender:   1,
		},
	}
	svc := newUserServiceFromDB(fdb)
	user, err := svc.GetProfile(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected non-nil user")
	}
	if user.ID != 1 {
		t.Errorf("expected ID=1, got %d", user.ID)
	}
	if user.Nickname != "Alice" {
		t.Errorf("expected Nickname=Alice, got %q", user.Nickname)
	}
	if fdb.lastFirstID != 1 {
		t.Errorf("expected DB queried with ID=1, got %d", fdb.lastFirstID)
	}
}

func TestGetProfile_NotFound(t *testing.T) {
	fdb := &fakeUserDB{
		firstError: gorm.ErrRecordNotFound,
	}
	svc := newUserServiceFromDB(fdb)
	_, err := svc.GetProfile(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for not found user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestGetProfile_DBError(t *testing.T) {
	fdb := &fakeUserDB{
		firstError: errors.New("connection refused"),
	}
	svc := newUserServiceFromDB(fdb)
	_, err := svc.GetProfile(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error when DB fails")
	}
	if !errors.Is(err, fdb.firstError) {
		t.Errorf("expected original error to be wrapped, got %v", err)
	}
}

// --- FindOne tests ---

func TestFindOne_Success(t *testing.T) {
	fdb := &fakeUserDB{
		firstUser: &User{
			Model:    gorm.Model{ID: 5},
			Openid:   "wx-openid-5",
			Nickname: "Bob",
			Avatar:   "https://example.com/bob.png",
		},
	}
	svc := newUserServiceFromDB(fdb)
	user, err := svc.FindOne(context.Background(), 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.ID != 5 {
		t.Errorf("expected ID=5, got %d", user.ID)
	}
	if user.Nickname != "Bob" {
		t.Errorf("expected Nickname=Bob, got %q", user.Nickname)
	}
}

func TestFindOne_NotFound(t *testing.T) {
	fdb := &fakeUserDB{
		firstError: gorm.ErrRecordNotFound,
	}
	svc := newUserServiceFromDB(fdb)
	_, err := svc.FindOne(context.Background(), 404)
	if err == nil {
		t.Fatal("expected error for not found user")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

// --- UpdateProfile tests ---

func TestUpdateProfile_Success_AllowedFields(t *testing.T) {
	existingUser := &User{
		Model:    gorm.Model{ID: 2},
		Openid:   "wx-openid-2",
		Nickname: "OldName",
		Avatar:   "https://example.com/old.png",
		Bio:      "old bio",
		Gender:   0,
	}
	updatedUser := &User{
		Model:    gorm.Model{ID: 2},
		Openid:   "wx-openid-2",
		Nickname: "NewName",
		Avatar:   "https://example.com/new.png",
		Bio:      "new bio",
		Gender:   1,
	}
	fdb := &fakeUserDB{
		firstUser: existingUser,
		saveUser:  updatedUser,
	}
	svc := newUserServiceFromDB(fdb)

	updates := map[string]interface{}{
		"nickname": "NewName",
		"avatar":   "https://example.com/new.png",
		"bio":      "new bio",
		"gender":   1,
	}

	user, err := svc.UpdateProfile(context.Background(), 2, updates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.Nickname != "NewName" {
		t.Errorf("expected updated Nickname=NewName, got %q", user.Nickname)
	}
}

func TestUpdateProfile_FiltersDeniedFields(t *testing.T) {
	existingUser := &User{
		Model:    gorm.Model{ID: 3},
		Openid:   "wx-openid-3",
		Nickname: "Charlie",
	}
	fdb := &fakeUserDB{
		firstUser: existingUser,
		saveUser:  existingUser,
	}
	svc := newUserServiceFromDB(fdb)

	// Try to update openid (not allowed) plus a valid field
	updates := map[string]interface{}{
		"openid":   "hacked-openid",
		"nickname": "Charlie Updated",
	}

	_, err := svc.UpdateProfile(context.Background(), 3, updates)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// The DB should NOT have received "openid" in the updates map
	if _, ok := fdb.lastUpdatesMap["openid"]; ok {
		t.Error("openid should be filtered out from updates")
	}
	// But nickname should pass through
	if _, ok := fdb.lastUpdatesMap["nickname"]; !ok {
		t.Error("nickname should be present in updates")
	}
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	fdb := &fakeUserDB{
		firstError: gorm.ErrRecordNotFound,
	}
	svc := newUserServiceFromDB(fdb)
	_, err := svc.UpdateProfile(context.Background(), 999, map[string]interface{}{"nickname": "X"})
	if err == nil {
		t.Fatal("expected error when user not found")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}

func TestUpdateProfile_DBSaveError(t *testing.T) {
	existingUser := &User{
		Model:    gorm.Model{ID: 4},
		Openid:   "wx-openid-4",
		Nickname: "Dave",
	}
	fdb := &fakeUserDB{
		firstUser: existingUser,
		saveError: errors.New("disk full"),
	}
	svc := newUserServiceFromDB(fdb)
	_, err := svc.UpdateProfile(context.Background(), 4, map[string]interface{}{"nickname": "Dave2"})
	if err == nil {
		t.Fatal("expected error when DB save fails")
	}
	if !errors.Is(err, fdb.saveError) {
		t.Errorf("expected save error to be wrapped, got %v", err)
	}
}

func TestUpdateProfile_EmptyUpdates_NoDBCall(t *testing.T) {
	existingUser := &User{
		Model:    gorm.Model{ID: 6},
		Openid:   "wx-openid-6",
		Nickname: "Eve",
	}
	fdb := &fakeUserDB{
		firstUser: existingUser,
		saveUser:  existingUser,
	}
	svc := newUserServiceFromDB(fdb)
	// All fields are denied/not in whitelist → filtered updates map is empty
	updates := map[string]interface{}{
		"openid": "bad",
		"id":     999,
	}
	user, err := svc.UpdateProfile(context.Background(), 6, updates)
	if err != nil {
		t.Fatalf("expected no error for empty allowed updates, got %v", err)
	}
	if user.Nickname != "Eve" {
		t.Errorf("expected unchanged nickname, got %q", user.Nickname)
	}
}
