package users

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// mockUserService is a test double for UserService.
type mockUserService struct {
	getProfileFunc              func(ctx context.Context, userID uint) (*User, error)
	findOneFunc                 func(ctx context.Context, userID uint) (*User, error)
	updateProfileFunc           func(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error)
	incrSubscribeFollowQuotaFunc func(ctx context.Context, userID uint) error
}

func (m *mockUserService) GetProfile(ctx context.Context, userID uint) (*User, error) {
	return m.getProfileFunc(ctx, userID)
}

func (m *mockUserService) FindOne(ctx context.Context, userID uint) (*User, error) {
	return m.findOneFunc(ctx, userID)
}

func (m *mockUserService) UpdateProfile(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
	return m.updateProfileFunc(ctx, userID, updates)
}

func (m *mockUserService) IncrSubscribeFollowQuota(ctx context.Context, userID uint) error {
	return m.incrSubscribeFollowQuotaFunc(ctx, userID)
}

func TestUserHandler_SubscribeFollow_Success(t *testing.T) {
	var gotUserID uint
	m := &mockUserService{
		incrSubscribeFollowQuotaFunc: func(ctx context.Context, userID uint) error {
			gotUserID = userID
			return nil
		},
	}
	h := &UserHandler{service: m}
	r := setupGin()
	r.POST("/users/subscribe/follow", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.SubscribeFollow(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/subscribe/follow", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotUserID != 42 {
		t.Errorf("expected quota incr for user 42, got %d", gotUserID)
	}
}

func TestUserHandler_SubscribeFollow_Unauthorized(t *testing.T) {
	m := &mockUserService{}
	h := &UserHandler{service: m}
	r := setupGin()
	r.POST("/users/subscribe/follow", h.SubscribeFollow) // no userID set

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/users/subscribe/follow", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func setupGin() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	mockUser := &User{
		Openid:   "openid123",
		Nickname: "testuser",
		Avatar:   "https://example.com/avatar.jpg",
		Bio:      "hello",
		Gender:   1,
	}

	svc := &mockUserService{
		getProfileFunc: func(ctx context.Context, userID uint) (*User, error) {
			if userID != 42 {
				t.Errorf("expected userID 42, got %d", userID)
			}
			return mockUser, nil
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "testuser") {
		t.Errorf("expected body to contain 'testuser', got %s", body)
	}
	if !strings.Contains(body, "success") {
		t.Errorf("expected body to contain 'success', got %s", body)
	}
}

func TestUserHandler_GetProfile_UserMissing_Unauthorized(t *testing.T) {
	svc := &mockUserService{
		getProfileFunc: func(ctx context.Context, userID uint) (*User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// token 有效但用户在库中不存在 → 登录态失效 → 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "login expired") {
		t.Errorf("expected body to contain 'login expired', got %s", body)
	}
}

func TestUserHandler_GetProfile_InternalError(t *testing.T) {
	svc := &mockUserService{
		getProfileFunc: func(ctx context.Context, userID uint) (*User, error) {
			return nil, errors.New("db error")
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.GetProfile(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "server error") {
		t.Errorf("expected body to contain 'server error', got %s", body)
	}
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	mockUser := &User{
		Openid:   "openid123",
		Nickname: "updated",
		Avatar:   "https://example.com/new.jpg",
		Bio:      "new bio",
		Gender:   2,
	}

	svc := &mockUserService{
		updateProfileFunc: func(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
			if userID != 42 {
				t.Errorf("expected userID 42, got %d", userID)
			}
			if updates["nickname"] != "updated" {
				t.Errorf("expected nickname 'updated', got %v", updates["nickname"])
			}
			return mockUser, nil
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.PATCH("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.UpdateProfile(c)
	})

	bodyStr := `{"nickname":"updated","avatar":"https://example.com/new.jpg","bio":"new bio","gender":2}`
	req := httptest.NewRequest(http.MethodPatch, "/users/profile", strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "updated") {
		t.Errorf("expected body to contain 'updated', got %s", body)
	}
}

func TestUserHandler_UpdateProfile_UserMissing_Unauthorized(t *testing.T) {
	svc := &mockUserService{
		updateProfileFunc: func(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.PATCH("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.UpdateProfile(c)
	})

	bodyStr := `{"nickname":"updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/users/profile", strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// token 有效但用户在库中不存在 → 登录态失效 → 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "login expired") {
		t.Errorf("expected body to contain 'login expired', got %s", body)
	}
}

func TestUserHandler_UpdateProfile_InternalError(t *testing.T) {
	svc := &mockUserService{
		updateProfileFunc: func(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
			return nil, errors.New("db error")
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.PATCH("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.UpdateProfile(c)
	})

	bodyStr := `{"nickname":"updated"}`
	req := httptest.NewRequest(http.MethodPatch, "/users/profile", strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestUserHandler_GetUser_Success(t *testing.T) {
	mockUser := &User{
		Openid:   "openid123",
		Nickname: "testuser",
		Avatar:   "https://example.com/avatar.jpg",
		Bio:      "hello",
		Gender:   1,
	}

	svc := &mockUserService{
		findOneFunc: func(ctx context.Context, userID uint) (*User, error) {
			if userID != 7 {
				t.Errorf("expected userID 7, got %d", userID)
			}
			return mockUser, nil
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/7", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !strings.Contains(body, "testuser") {
		t.Errorf("expected body to contain 'testuser', got %s", body)
	}
}

func TestUserHandler_GetUser_NotFound(t *testing.T) {
	svc := &mockUserService{
		findOneFunc: func(ctx context.Context, userID uint) (*User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "user not found") {
		t.Errorf("expected body to contain 'user not found', got %s", body)
	}
}

func TestUserHandler_GetUser_InvalidID(t *testing.T) {
	svc := &mockUserService{}
	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUserHandler_UpdateProfile_FiltersNonWhitelistedFields(t *testing.T) {
	mockUser := &User{
		Openid:   "openid123",
		Nickname: "updated",
		Avatar:   "https://example.com/avatar.jpg",
		Bio:      "",
		Gender:   0,
	}

	svc := &mockUserService{
		updateProfileFunc: func(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error) {
			if userID != 42 {
				t.Errorf("expected userID 42, got %d", userID)
			}
			if _, ok := updates["password"]; ok {
				t.Errorf("expected password to be filtered out, got %v", updates["password"])
			}
			if _, ok := updates["openid"]; ok {
				t.Errorf("expected openid to be filtered out, got %v", updates["openid"])
			}
			if updates["nickname"] != "updated" {
				t.Errorf("expected nickname 'updated', got %v", updates["nickname"])
			}
			return mockUser, nil
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.PATCH("/users/profile", func(c *gin.Context) {
		c.Set("userID", uint(42))
		h.UpdateProfile(c)
	})

	bodyStr := `{"nickname":"updated","password":"secret","openid":"evil"}`
	req := httptest.NewRequest(http.MethodPatch, "/users/profile", strings.NewReader(bodyStr))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "updated") {
		t.Errorf("expected body to contain 'updated', got %s", body)
	}
}

func TestUserHandler_GetProfile_Unauthorized(t *testing.T) {
	svc := &mockUserService{}
	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/profile", h.GetProfile)

	req := httptest.NewRequest(http.MethodGet, "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "unauthorized") {
		t.Errorf("expected body to contain 'unauthorized', got %s", body)
	}
}

func TestUserHandler_UpdateProfile_Unauthorized(t *testing.T) {
	svc := &mockUserService{}
	h := &UserHandler{service: svc}
	r := setupGin()
	r.PATCH("/users/profile", h.UpdateProfile)

	req := httptest.NewRequest(http.MethodPatch, "/users/profile", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "unauthorized") {
		t.Errorf("expected body to contain 'unauthorized', got %s", body)
	}
}

func TestUserHandler_GetUser_InternalError(t *testing.T) {
	svc := &mockUserService{
		findOneFunc: func(ctx context.Context, userID uint) (*User, error) {
			return nil, errors.New("db error")
		},
	}

	h := &UserHandler{service: svc}
	r := setupGin()
	r.GET("/users/:id", h.GetUser)

	req := httptest.NewRequest(http.MethodGet, "/users/7", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
