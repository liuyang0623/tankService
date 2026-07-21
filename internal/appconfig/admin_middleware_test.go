package appconfig

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// fakeOpenidLookup 实现 openidLookup,供中间件测试注入,无需真实 DB。
type fakeOpenidLookup struct {
	openid string
	err    error
	lastID uint
}

func (f *fakeOpenidLookup) OpenidByUserID(userID uint) (string, error) {
	f.lastID = userID
	return f.openid, f.err
}

// setupAdminRoute 构造一个先注入 userID、再挂管理员中间件的测试引擎。
// injectUserID=0 时不注入，模拟上游 JWT 缺失。
func setupAdminRoute(lookup openidLookup, whitelist []string, injectUserID uint) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if injectUserID != 0 {
			c.Set("userID", injectUserID)
		}
		c.Next()
	})
	r.PUT("/app-config", newAdminOnly(lookup, whitelist), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// TestAdminOnly_Whitelisted_Passes 白名单命中放行（handler 执行，返回 200）。
func TestAdminOnly_Whitelisted_Passes(t *testing.T) {
	lookup := &fakeOpenidLookup{openid: "admin-openid"}
	r := setupAdminRoute(lookup, []string{"admin-openid", "other"}, 7)

	req := httptest.NewRequest("PUT", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if lookup.lastID != 7 {
		t.Errorf("expected lookup by userID=7, got %d", lookup.lastID)
	}
}

// TestAdminOnly_NotWhitelisted_Returns403 非白名单用户被拒（403，handler 不执行）。
func TestAdminOnly_NotWhitelisted_Returns403(t *testing.T) {
	lookup := &fakeOpenidLookup{openid: "stranger-openid"}
	r := setupAdminRoute(lookup, []string{"admin-openid"}, 9)

	req := httptest.NewRequest("PUT", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// TestAdminOnly_NoUserID_Returns403 上游未注入 userID 时直接 403（fail-closed）。
func TestAdminOnly_NoUserID_Returns403(t *testing.T) {
	lookup := &fakeOpenidLookup{openid: "admin-openid"}
	r := setupAdminRoute(lookup, []string{"admin-openid"}, 0)

	req := httptest.NewRequest("PUT", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}

// TestAdminOnly_UserNotFound_Returns403 反查用户不存在时返回 403（而非 500）。
func TestAdminOnly_UserNotFound_Returns403(t *testing.T) {
	lookup := &fakeOpenidLookup{err: gorm.ErrRecordNotFound}
	r := setupAdminRoute(lookup, []string{"admin-openid"}, 5)

	req := httptest.NewRequest("PUT", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
	}
}
