package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockAuthService implements authServiceIface for testing.
type mockAuthService struct {
	result *LoginResult
	err    error
}

func (m *mockAuthService) WechatLogin(_ context.Context, _, _, _ string) (*LoginResult, error) {
	return m.result, m.err
}

func setupRouter(handler *AuthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/v1/auth/wechat/login", handler.WechatLogin)
	return r
}

func TestWechatLogin_MissingCode(t *testing.T) {
	svc := &mockAuthService{}
	h := NewAuthHandler(svc)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] == nil {
		t.Error("expected 'message' field in response")
	}
	if resp["code"] == nil {
		t.Error("expected 'code' field in response")
	}
}

func TestWechatLogin_EmptyCode(t *testing.T) {
	svc := &mockAuthService{}
	h := NewAuthHandler(svc)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"code": ""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestWechatLogin_ServiceError(t *testing.T) {
	svc := &mockAuthService{err: errors.New("wechat API failed")}
	h := NewAuthHandler(svc)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"code": "valid_code"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] == nil {
		t.Error("expected 'message' field in response")
	}
	if resp["code"] == nil {
		t.Error("expected 'code' field in response")
	}
}

func TestWechatLogin_Success(t *testing.T) {
	svc := &mockAuthService{result: &LoginResult{
		Token: "jwt_token_here",
		User: LoginUserInfo{
			ID:       1,
			Nickname: "test_user",
			Avatar:   "https://example.com/avatar.png",
		},
	}}
	h := NewAuthHandler(svc)
	r := setupRouter(h)

	body := bytes.NewBufferString(`{"code": "wx_code_here", "nickName": "test_user", "avatarUrl": "https://example.com/avatar.png"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp["message"] != "success" {
		t.Errorf("expected message 'success', got %v", resp["message"])
	}
	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'data' to be an object, got %T", resp["data"])
	}
	if data["token"] != "jwt_token_here" {
		t.Errorf("expected token 'jwt_token_here', got %v", data["token"])
	}
	user, ok := data["user"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected 'user' to be an object, got %T", data["user"])
	}
	if user["nickname"] != "test_user" {
		t.Errorf("expected nickname 'test_user', got %v", user["nickname"])
	}
}
