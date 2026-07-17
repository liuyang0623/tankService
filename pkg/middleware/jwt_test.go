package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

const testSecret = "test-secret-key"

func setupTestRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", JWTMiddleware(secret), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})
	return r
}

func TestJWTMiddleware_MissingAuthorizationHeader(t *testing.T) {
	r := setupTestRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "authorization header required" {
		t.Errorf("expected message 'authorization header required', got %v", body["message"])
	}
	if int(body["code"].(float64)) != 401 {
		t.Errorf("expected code 401, got %v", body["code"])
	}
}

func TestJWTMiddleware_InvalidFormat_NoBearer(t *testing.T) {
	r := setupTestRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "invalid-token")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "invalid authorization format" {
		t.Errorf("expected message 'invalid authorization format', got %v", body["message"])
	}
}

func TestJWTMiddleware_InvalidFormat_BearerOnly(t *testing.T) {
	r := setupTestRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "invalid authorization format" {
		t.Errorf("expected message 'invalid authorization format', got %v", body["message"])
	}
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	r := setupTestRouter(testSecret)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer this.is.not.valid")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "invalid or expired token" {
		t.Errorf("expected message 'invalid or expired token', got %v", body["message"])
	}
}

func TestJWTMiddleware_WrongSecret(t *testing.T) {
	r := setupTestRouter(testSecret)

	// Generate a token with a different secret
	token, err := GenerateToken(42, "wrong-secret")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "invalid or expired token" {
		t.Errorf("expected message 'invalid or expired token', got %v", body["message"])
	}
}

func TestJWTMiddleware_ValidToken(t *testing.T) {
	r := setupTestRouter(testSecret)

	var expectedUserID uint = 123
	token, err := GenerateToken(expectedUserID, testSecret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if uint(body["user_id"].(float64)) != expectedUserID {
		t.Errorf("expected user_id %d, got %v", expectedUserID, body["user_id"])
	}
}

func TestGenerateToken(t *testing.T) {
	var userID uint = 99
	token, err := GenerateToken(userID, testSecret)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Error("GenerateToken returned empty token")
	}
}
