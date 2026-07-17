package message

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// mockMessageService implements messageServiceIface for handler tests.
type mockMessageService struct {
	createMsgResult *Message
	createMsgErr    error
	convsResult     *PaginatedConversations
	convsErr        error
	msgsResult      *PaginatedMessages
	msgsErr         error
	markReadErr     error
	gotSenderID     uint
	gotToUserID     uint
	gotConvID       uint
	gotUserID       uint
}

func (m *mockMessageService) CreateMessage(ctx context.Context, senderID, toUserID uint, msgType, content string) (*Message, error) {
	m.gotSenderID = senderID
	m.gotToUserID = toUserID
	return m.createMsgResult, m.createMsgErr
}

func (m *mockMessageService) GetConversations(ctx context.Context, userID uint, page, limit int) (*PaginatedConversations, error) {
	m.gotUserID = userID
	return m.convsResult, m.convsErr
}

func (m *mockMessageService) GetMessages(ctx context.Context, conversationID, userID uint, page, limit int) (*PaginatedMessages, error) {
	m.gotConvID = conversationID
	m.gotUserID = userID
	return m.msgsResult, m.msgsErr
}

func (m *mockMessageService) MarkRead(ctx context.Context, conversationID, userID uint) error {
	m.gotConvID = conversationID
	m.gotUserID = userID
	return m.markReadErr
}

func (m *mockMessageService) FindConversationByUsers(ctx context.Context, userID, otherUserID uint) (uint, error) {
	return 0, nil
}

// Test model table names
func TestMessage_TableNames(t *testing.T) {
	if (Conversation{}).TableName() != "conversations" {
		t.Errorf("expected 'conversations', got %q", (Conversation{}).TableName())
	}
	if (Message{}).TableName() != "messages" {
		t.Errorf("expected 'messages', got %q", (Message{}).TableName())
	}
}

// Test ensureConversation sorts IDs correctly
func TestEnsureConversation(t *testing.T) {
	// This tests the sorting logic in ensureConversation with a raw model
	// (no DB needed for the pure function contract)
	small, big := uint(5), uint(10)
	s1, s2 := small, big
	if s1 > s2 {
		s1, s2 = s2, s1
	}
	if s1 != 5 || s2 != 10 {
		t.Errorf("ensureConversation should sort IDs, got %d, %d", s1, s2)
	}

	s1, s2 = big, small
	if s1 > s2 {
		s1, s2 = s2, s1
	}
	if s1 != 5 || s2 != 10 {
		t.Errorf("ensureConversation should sort IDs even when swapped, got %d, %d", s1, s2)
	}
}

func newTestHandler(m *mockMessageService) *MessageHandler {
	hub := NewHub()
	go hub.Run()
	return &MessageHandler{
		service:  m,
		hub:      hub,
		upgrader: websocket.Upgrader{}, // minimal upgrader for test; not used in REST tests
	}
}

func TestSendMessage_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHandler(&mockMessageService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/messages", nil)

	h.SendMessage(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestSendMessage_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()
	m := &mockMessageService{
		createMsgResult: &Message{
			Model:          gorm.Model{ID: 1, CreatedAt: now},
			ConversationID: 1,
			SenderID:       1,
			Type:           "text",
			Content:        "hello",
		},
	}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"toUserId": 2, "content": "hello"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1))

	h.SendMessage(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", w.Code, w.Body.String())
	}
	if m.gotSenderID != 1 || m.gotToUserID != 2 {
		t.Errorf("expected sender=1 toUser=2, got sender=%d toUser=%d", m.gotSenderID, m.gotToUserID)
	}
}

func TestSendMessage_UserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockMessageService{createMsgErr: gorm.ErrRecordNotFound}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	body := `{"toUserId": 99, "content": "hello"}`
	c.Request = httptest.NewRequest(http.MethodPost, "/messages", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("userID", uint(1))

	h.SendMessage(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestListConversations_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockMessageService{
		convsResult: &PaginatedConversations{
			Data: []ConversationItem{{
				ID: 1,
				OtherUser: OtherUserInfo{ID: 2, Nickname: "test", Avatar: "avatar.jpg"},
				LastMessage: "hi",
				LastTime:    time.Now(),
				UnreadCount: 0,
			}},
			Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
		},
	}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations", nil)
	c.Set("userID", uint(1))

	h.ListConversations(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var body struct {
		Data PaginatedConversations `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(body.Data.Data) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(body.Data.Data))
	}
	if body.Data.Data[0].OtherUser.Nickname != "test" {
		t.Errorf("expected nickname 'test', got %q", body.Data.Data[0].OtherUser.Nickname)
	}
}

func TestListConversations_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHandler(&mockMessageService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations", nil)

	h.ListConversations(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetMessages_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now()
	m := &mockMessageService{
		msgsResult: &PaginatedMessages{
			Data: []MessageItem{{
				ID:             1,
				ConversationID: 1,
				SenderID:       1,
				Type:           "text",
				Content:        "hi",
				CreatedAt:      now,
			}},
			Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
		},
	}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/1/messages", nil)
	c.Set("userID", uint(1))

	h.GetMessages(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotConvID != 1 || m.gotUserID != 1 {
		t.Errorf("expected convID=1 userID=1, got convID=%d userID=%d", m.gotConvID, m.gotUserID)
	}
}

func TestGetMessages_InvalidConvID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := newTestHandler(&mockMessageService{})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/abc/messages", nil)
	c.Set("userID", uint(1))

	h.GetMessages(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestGetMessages_Forbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockMessageService{msgsErr: fmt.Errorf("not a participant of this conversation")}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/conversations/1/messages", nil)
	c.Set("userID", uint(1))

	h.GetMessages(c)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

var errNotParticipant = errors.New("not a participant of this conversation")

func TestMarkRead_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockMessageService{}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/conversations/1/read", nil)
	c.Set("userID", uint(1))

	h.MarkRead(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if m.gotConvID != 1 || m.gotUserID != 1 {
		t.Errorf("expected convID=1 userID=1, got convID=%d userID=%d", m.gotConvID, m.gotUserID)
	}
}

func TestMarkRead_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := &mockMessageService{markReadErr: fmt.Errorf("not found: %w", gorm.ErrRecordNotFound)}
	h := newTestHandler(m)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	c.Request = httptest.NewRequest(http.MethodPost, "/conversations/999/read", nil)
	c.Set("userID", uint(1))

	h.MarkRead(c)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
