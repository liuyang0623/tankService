package follow

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// mockFollowService implements followServiceIface for handler tests.
type mockFollowService struct {
	toggleFollowing bool
	toggleErr       error
	listResult      *PaginatedUsers
	listErr         error
	gotFollowerID   uint
	gotTargetID     uint
	gotCurrentID    uint
}

func (m *mockFollowService) ToggleFollow(ctx context.Context, followerID, targetID uint) (bool, error) {
	m.gotFollowerID = followerID
	m.gotTargetID = targetID
	return m.toggleFollowing, m.toggleErr
}

func (m *mockFollowService) ListFollowers(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error) {
	m.gotTargetID = userID
	m.gotCurrentID = currentUserID
	return m.listResult, m.listErr
}

func (m *mockFollowService) ListFollowing(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error) {
	m.gotTargetID = userID
	m.gotCurrentID = currentUserID
	return m.listResult, m.listErr
}

func newHandlerWith(m *mockFollowService) *FollowHandler {
	return &FollowHandler{service: m}
}

func TestFollow_TableName(t *testing.T) {
	if (Follow{}).TableName() != "follows" {
		t.Errorf("expected table name 'follows', got %q", (Follow{}).TableName())
	}
}

func TestToggleFollow_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWith(&mockFollowService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/users/2/follow", nil)
	// no userID set in context

	h.ToggleFollow(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestToggleFollow_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockFollowService{toggleFollowing: true}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "2"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/users/2/follow", nil)
	c.Set("userID", uint(1))

	h.ToggleFollow(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotFollowerID != 1 || m.gotTargetID != 2 {
		t.Errorf("expected follower=1 target=2, got follower=%d target=%d", m.gotFollowerID, m.gotTargetID)
	}
	var body struct {
		Data struct {
			Following bool `json:"following"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if !body.Data.Following {
		t.Errorf("expected following=true in response")
	}
}

func TestToggleFollow_SelfNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// service returns record-not-found -> handler maps to 404
	m := &mockFollowService{toggleErr: gorm.ErrRecordNotFound}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "99"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/users/99/follow", nil)
	c.Set("userID", uint(1))

	h.ToggleFollow(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestToggleFollow_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWith(&mockFollowService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/users/abc/follow", nil)
	c.Set("userID", uint(1))

	h.ToggleFollow(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListFollowers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockFollowService{listResult: &PaginatedUsers{
		Data: []FollowUserItem{{ID: 5, Nickname: "a", IsFollowing: true}},
		Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
	}}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/users/3/followers", nil)
	c.Set("userID", uint(7)) // optional auth present

	h.ListFollowers(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotTargetID != 3 || m.gotCurrentID != 7 {
		t.Errorf("expected target=3 current=7, got target=%d current=%d", m.gotTargetID, m.gotCurrentID)
	}
}

func TestListFollowing_OptionalAuthAbsent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockFollowService{listResult: &PaginatedUsers{Data: []FollowUserItem{}, Meta: PaginationMeta{}}}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "3"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/users/3/following", nil)
	// no userID -> currentUserID should be 0

	h.ListFollowing(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotCurrentID != 0 {
		t.Errorf("expected currentUserID=0 when not logged in, got %d", m.gotCurrentID)
	}
}
