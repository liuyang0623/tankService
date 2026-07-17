package diary

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// mockDiaryService implements diaryServiceIface for handler tests.
type mockDiaryService struct {
	createFn  func(ctx context.Context, userID uint, input CreateDiaryInput) (*DiaryResponse, error)
	findMineFn func(ctx context.Context, userID uint, notebookID uint, page, limit int) (*PaginatedResult, error)
	findOneFn func(ctx context.Context, id, userID uint) (*DiaryResponse, error)
	updateFn  func(ctx context.Context, id, userID uint, input UpdateDiaryInput) (*DiaryResponse, error)
	removeFn  func(ctx context.Context, id, userID uint) error
}

func (m *mockDiaryService) Create(ctx context.Context, userID uint, input CreateDiaryInput) (*DiaryResponse, error) {
	return m.createFn(ctx, userID, input)
}
func (m *mockDiaryService) FindMine(ctx context.Context, userID uint, notebookID uint, page, limit int) (*PaginatedResult, error) {
	return m.findMineFn(ctx, userID, notebookID, page, limit)
}
func (m *mockDiaryService) FindOne(ctx context.Context, id, userID uint) (*DiaryResponse, error) {
	return m.findOneFn(ctx, id, userID)
}
func (m *mockDiaryService) Update(ctx context.Context, id, userID uint, input UpdateDiaryInput) (*DiaryResponse, error) {
	return m.updateFn(ctx, id, userID, input)
}
func (m *mockDiaryService) Remove(ctx context.Context, id, userID uint) error {
	return m.removeFn(ctx, id, userID)
}

func setupTest() (*gin.Engine, *mockDiaryService) {
	gin.SetMode(gin.TestMode)
	mock := &mockDiaryService{}
	h := &DiaryHandler{service: mock}
	r := gin.New()
	// Simulate JWTMiddleware: inject userID = 1
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(1))
	})
	r.POST("/diaries", h.Create)
	r.GET("/diaries", h.FindMine)
	r.GET("/diaries/:id", h.FindOne)
	r.PATCH("/diaries/:id", h.Update)
	r.DELETE("/diaries/:id", h.Remove)
	return r, mock
}

func TestCreate_Success(t *testing.T) {
	r, mock := setupTest()
	now := time.Now()
	mock.createFn = func(ctx context.Context, userID uint, input CreateDiaryInput) (*DiaryResponse, error) {
		if userID != 1 {
			t.Errorf("expected userID=1, got %d", userID)
		}
		return &DiaryResponse{
			ID:        1,
			Title:     input.Title,
			Content:   input.Content,
			Cover:     input.Cover,
			Mood:      input.Mood,
			Weather:   input.Weather,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	body := `{"title":"Test Diary","content":"Hello world","mood":"happy","weather":"sunny"}`
	req := httptest.NewRequest("POST", "/diaries", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "Test Diary" {
		t.Errorf("expected title 'Test Diary', got %v", data["title"])
	}
	if data["mood"] != "happy" {
		t.Errorf("expected mood 'happy', got %v", data["mood"])
	}
	if data["weather"] != "sunny" {
		t.Errorf("expected weather 'sunny', got %v", data["weather"])
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	r, _ := setupTest()
	body := `{"title":"","content":"Hello"}`
	req := httptest.NewRequest("POST", "/diaries", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty title, got %d", w.Code)
	}
}

func TestCreate_EmptyContent(t *testing.T) {
	r, _ := setupTest()
	body := `{"title":"Hi","content":""}`
	req := httptest.NewRequest("POST", "/diaries", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty content, got %d", w.Code)
	}
}

func TestFindMine_Success(t *testing.T) {
	r, mock := setupTest()
	now := time.Now()
	mock.findMineFn = func(ctx context.Context, userID uint, notebookID uint, page, limit int) (*PaginatedResult, error) {
		return &PaginatedResult{
			Data: []DiaryListItem{
				{ID: 1, Title: "Diary 1", ContentPreview: "Hello...", Mood: "happy", CreatedAt: now},
			},
			Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
		}, nil
	}

	req := httptest.NewRequest("GET", "/diaries", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	items := data["data"].([]interface{})
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestFindOne_Success(t *testing.T) {
	r, mock := setupTest()
	now := time.Now()
	mock.findOneFn = func(ctx context.Context, id, userID uint) (*DiaryResponse, error) {
		return &DiaryResponse{
			ID:        id,
			Title:     "My Diary",
			Content:   "Full content here",
			Mood:      "happy",
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	req := httptest.NewRequest("GET", "/diaries/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestFindOne_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.findOneFn = func(ctx context.Context, id, userID uint) (*DiaryResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}

	req := httptest.NewRequest("GET", "/diaries/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUpdate_Success(t *testing.T) {
	r, mock := setupTest()
	now := time.Now()
	mock.updateFn = func(ctx context.Context, id, userID uint, input UpdateDiaryInput) (*DiaryResponse, error) {
		return &DiaryResponse{
			ID:        id,
			Title:     *input.Title,
			Content:   *input.Content,
			Mood:      *input.Mood,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	body := `{"title":"Updated","content":"New content","mood":"sad"}`
	req := httptest.NewRequest("PATCH", "/diaries/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "Updated" {
		t.Errorf("expected title 'Updated', got %v", data["title"])
	}
	if data["mood"] != "sad" {
		t.Errorf("expected mood 'sad', got %v", data["mood"])
	}
}

func TestRemove_Success(t *testing.T) {
	r, mock := setupTest()
	mock.removeFn = func(ctx context.Context, id, userID uint) error {
		return nil
	}

	req := httptest.NewRequest("DELETE", "/diaries/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRemove_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.removeFn = func(ctx context.Context, id, userID uint) error {
		return gorm.ErrRecordNotFound
	}

	req := httptest.NewRequest("DELETE", "/diaries/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUnauthorized_NoJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &DiaryHandler{service: &mockDiaryService{}}
	r := gin.New()
	// No JWT middleware — userID not set
	r.POST("/diaries", h.Create)

	body := `{"title":"X","content":"Y"}`
	req := httptest.NewRequest("POST", "/diaries", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}