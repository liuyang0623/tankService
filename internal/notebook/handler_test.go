package notebook

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

// mockNotebookService 实现 notebookServiceIface 供 handler 测试。
type mockNotebookService struct {
	createFn   func(ctx context.Context, userID uint, input CreateNotebookInput) (*NotebookResponse, error)
	findMineFn func(ctx context.Context, userID uint) ([]NotebookResponse, error)
	updateFn   func(ctx context.Context, id, userID uint, input UpdateNotebookInput) (*NotebookResponse, error)
	removeFn   func(ctx context.Context, id, userID uint) error
}

func (m *mockNotebookService) Create(ctx context.Context, userID uint, input CreateNotebookInput) (*NotebookResponse, error) {
	return m.createFn(ctx, userID, input)
}
func (m *mockNotebookService) FindMine(ctx context.Context, userID uint) ([]NotebookResponse, error) {
	return m.findMineFn(ctx, userID)
}
func (m *mockNotebookService) Update(ctx context.Context, id, userID uint, input UpdateNotebookInput) (*NotebookResponse, error) {
	return m.updateFn(ctx, id, userID, input)
}
func (m *mockNotebookService) Remove(ctx context.Context, id, userID uint) error {
	return m.removeFn(ctx, id, userID)
}

func setupTest() (*gin.Engine, *mockNotebookService) {
	gin.SetMode(gin.TestMode)
	mock := &mockNotebookService{}
	h := &NotebookHandler{service: mock}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userID", uint(1)) })
	r.POST("/notebooks", h.Create)
	r.GET("/notebooks", h.FindMine)
	r.PATCH("/notebooks/:id", h.Update)
	r.DELETE("/notebooks/:id", h.Remove)
	return r, mock
}

func TestCreate_Success(t *testing.T) {
	r, mock := setupTest()
	mock.createFn = func(ctx context.Context, userID uint, input CreateNotebookInput) (*NotebookResponse, error) {
		return &NotebookResponse{ID: 1, Name: input.Name, Color: input.Color, CreatedAt: time.Now()}, nil
	}
	req := httptest.NewRequest("POST", "/notebooks", strings.NewReader(`{"name":"旅行","color":"#a6c0ce"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["name"] != "旅行" {
		t.Errorf("expected name 旅行, got %v", data["name"])
	}
}

func TestCreate_EmptyName(t *testing.T) {
	r, _ := setupTest()
	req := httptest.NewRequest("POST", "/notebooks", strings.NewReader(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestFindMine_Success(t *testing.T) {
	r, mock := setupTest()
	mock.findMineFn = func(ctx context.Context, userID uint) ([]NotebookResponse, error) {
		return []NotebookResponse{{ID: 1, Name: "默认", DiaryCount: 3}}, nil
	}
	req := httptest.NewRequest("GET", "/notebooks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.updateFn = func(ctx context.Context, id, userID uint, input UpdateNotebookInput) (*NotebookResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	req := httptest.NewRequest("PATCH", "/notebooks/999", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRemove_Success(t *testing.T) {
	r, mock := setupTest()
	mock.removeFn = func(ctx context.Context, id, userID uint) error { return nil }
	req := httptest.NewRequest("DELETE", "/notebooks/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestRemove_NotFound(t *testing.T) {
	r, mock := setupTest()
	mock.removeFn = func(ctx context.Context, id, userID uint) error { return gorm.ErrRecordNotFound }
	req := httptest.NewRequest("DELETE", "/notebooks/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestUnauthorized_NoJWT(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &NotebookHandler{service: &mockNotebookService{}}
	r := gin.New()
	r.POST("/notebooks", h.Create)
	req := httptest.NewRequest("POST", "/notebooks", strings.NewReader(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
