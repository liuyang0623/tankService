package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-service/pkg/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestSetupRouter_NotNil verifies that setupRouter returns a non-nil engine.
func TestSetupRouter_NotNil(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
	}
	r := setupRouter(nil, cfg)
	if r == nil {
		t.Fatal("setupRouter returned nil")
	}
}

// TestSetupRouter_CORS verifies that CORS headers are present on OPTIONS preflight requests.
// The request URL uses a distinct host (myserver.local) so the gin-contrib/cors middleware
// does not treat the request as same-origin and correctly applies preflight handling.
func TestSetupRouter_CORS(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "http://example.com",
	}
	r := setupRouter(nil, cfg)

	// Use a full URL with a different host so gin-contrib/cors treats this as
	// a cross-origin request (not same-origin).
	req := httptest.NewRequest(http.MethodOptions, "http://myserver.local/api/v1/health", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin == "" {
		t.Errorf("expected Access-Control-Allow-Origin header to be set, got empty string (status=%d)", w.Code)
	}
	if allowOrigin != "http://example.com" {
		t.Errorf("expected Access-Control-Allow-Origin=http://example.com, got %q", allowOrigin)
	}
}

// TestSetupRouter_CORS_WildcardOrigin verifies that wildcard CORSOrigin config sets AllowAllOrigins.
func TestSetupRouter_CORS_WildcardOrigin(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
	}
	r := setupRouter(nil, cfg)

	req := httptest.NewRequest(http.MethodGet, "http://myserver.local/api/v1/", nil)
	req.Header.Set("Origin", "http://any-origin.com")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// With AllowAllOrigins, the header should be "*"
	allowOrigin := w.Header().Get("Access-Control-Allow-Origin")
	if allowOrigin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin=*, got %q (status=%d)", allowOrigin, w.Code)
	}
}

// TestSetupRouter_APIPrefix verifies the /api/v1 route group is registered without errors.
func TestSetupRouter_APIPrefix(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
	}
	r := setupRouter(nil, cfg)

	// The route group itself has no handlers yet; a GET returns 404.
	// What we must NOT see is a 500 or a panic.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusInternalServerError {
		t.Errorf("unexpected 500 on /api/v1/: %d", w.Code)
	}
}

// TestSetupRouter_CustomAPIPrefix verifies a non-default prefix is respected.
func TestSetupRouter_CustomAPIPrefix(t *testing.T) {
	cfg := &config.Config{
		Port:       "8080",
		APIPrefix:  "v2",
		CORSOrigin: "*",
	}
	r := setupRouter(nil, cfg)

	req := httptest.NewRequest(http.MethodGet, "/v2/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusInternalServerError {
		t.Errorf("unexpected 500 on /v2/: %d", w.Code)
	}
}

// TestSetupRouter_SwaggerUI verifies that /api/docs/index.html returns a non-404 response.
func TestSetupRouter_SwaggerUI(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
	}
	r := setupRouter(nil, cfg)

	req := httptest.NewRequest(http.MethodGet, "/api/docs/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("expected /api/docs/index.html to be accessible (non-404), got 404")
	}
}

// TestSetupRouter_WechatLoginRouteRegistered verifies that POST /api/v1/auth/wechat/login
// is registered and returns a non-404 response (route exists, even if request body is invalid).
func TestSetupRouter_WechatLoginRouteRegistered(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
		JWTSecret:  "test-secret",
	}
	var db *gorm.DB // nil — route registration doesn't require a real DB connection
	r := setupRouter(db, cfg)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/wechat/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Errorf("expected POST /api/v1/auth/wechat/login to be registered (non-404), got 404")
	}
}

// TestSetupRouter_UserRoutesRegistered verifies that user routes are registered.
func TestSetupRouter_UserRoutesRegistered(t *testing.T) {
	cfg := &config.Config{
		Port:       "3000",
		APIPrefix:  "api/v1",
		CORSOrigin: "*",
		JWTSecret:  "test-secret",
	}
	var db *gorm.DB
	r := setupRouter(db, cfg)

	routes := r.Routes()
	found := false
	for _, route := range routes {
		if route.Method == "GET" && route.Path == "/api/v1/users/:id" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected GET /api/v1/users/:id to be registered")
	}
}
