package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockNotificationService implements notificationServiceIface for handler tests.
type mockNotificationService struct {
	listResult *PaginatedNotifications
	listErr    error
	summary    *UnreadSummary
	summaryErr error
	markErr    error
	gotUserID  uint
	markCalled bool
}

func (m *mockNotificationService) List(ctx context.Context, userID uint, page, limit int) (*PaginatedNotifications, error) {
	m.gotUserID = userID
	return m.listResult, m.listErr
}

func (m *mockNotificationService) MarkAllRead(ctx context.Context, userID uint) error {
	m.gotUserID = userID
	m.markCalled = true
	return m.markErr
}

func (m *mockNotificationService) UnreadSummary(ctx context.Context, userID uint) (*UnreadSummary, error) {
	m.gotUserID = userID
	return m.summary, m.summaryErr
}

func newHandlerWith(m *mockNotificationService) *NotificationHandler {
	return &NotificationHandler{service: m}
}

func withUser(c *gin.Context, uid uint) {
	c.Set("userID", uid)
}

func TestNotification_TableName(t *testing.T) {
	if (Notification{}).TableName() != "notifications" {
		t.Errorf("expected 'notifications', got %q", (Notification{}).TableName())
	}
}

func TestList_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWith(&mockNotificationService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notifications", nil)
	// no userID in context
	h.List(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockNotificationService{listResult: &PaginatedNotifications{
		Data: []NotificationItem{{ID: 1, Type: "follow", Actor: ActorInfo{ID: 9, Nickname: "b"}}},
		Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
	}}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notifications", nil)
	withUser(c, 42)
	h.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotUserID != 42 {
		t.Errorf("expected service called with userID 42, got %d", m.gotUserID)
	}
}

func TestMarkRead_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockNotificationService{}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/notifications/read", nil)
	withUser(c, 7)
	h.MarkRead(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !m.markCalled || m.gotUserID != 7 {
		t.Errorf("expected MarkAllRead called for user 7, got called=%v user=%d", m.markCalled, m.gotUserID)
	}
}

func TestUnreadCount_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockNotificationService{summary: &UnreadSummary{UnreadCount: 3, Latest: &NotificationItem{ID: 5, Type: "follow"}}}
	h := newHandlerWith(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	withUser(c, 1)
	h.UnreadCount(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Data UnreadSummary `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Data.UnreadCount != 3 {
		t.Errorf("expected unreadCount 3, got %d", resp.Data.UnreadCount)
	}
}

func TestUnreadCount_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newHandlerWith(&mockNotificationService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/notifications/unread-count", nil)
	h.UnreadCount(c)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
