package appconfig

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// mockAppConfigService 实现 appConfigServiceIface 供 handler 测试。
type mockAppConfigService struct {
	getFn    func(ctx context.Context) (*AppConfigResponse, error)
	updateFn func(ctx context.Context, auditMode bool) error
}

func (m *mockAppConfigService) GetConfig(ctx context.Context) (*AppConfigResponse, error) {
	return m.getFn(ctx)
}

func (m *mockAppConfigService) UpdateAuditMode(ctx context.Context, auditMode bool) error {
	return m.updateFn(ctx, auditMode)
}

func setupTest(mock *mockAppConfigService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &AppConfigHandler{service: mock}
	r := gin.New()
	// 注意：不挂任何 JWT 中间件，模拟免鉴权公开路由
	r.GET("/app-config", h.GetConfig)
	return r
}

// TestGetConfig_NoAuth_Returns200 验证免鉴权即可访问（无 JWT 返回 200 而非 401）。
func TestGetConfig_NoAuth_Returns200(t *testing.T) {
	r := setupTest(&mockAppConfigService{
		getFn: func(ctx context.Context) (*AppConfigResponse, error) {
			return &AppConfigResponse{AuditMode: true}, nil
		},
	})
	req := httptest.NewRequest("GET", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

// TestGetConfig_ResponseShape 验证响应体形状为 {data:{auditMode:...}, code:200, message:"success"}
// 且字段名为 camelCase auditMode。
func TestGetConfig_ResponseShape(t *testing.T) {
	r := setupTest(&mockAppConfigService{
		getFn: func(ctx context.Context) (*AppConfigResponse, error) {
			return &AppConfigResponse{AuditMode: true}, nil
		},
	})
	req := httptest.NewRequest("GET", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var body struct {
		Data    map[string]interface{} `json:"data"`
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v; raw=%s", err, w.Body.String())
	}
	if body.Code != 200 || body.Message != "success" {
		t.Errorf("expected code=200 message=success, got code=%d message=%s", body.Code, body.Message)
	}
	// camelCase 字段名断言：必须是 auditMode，不能是 audit_mode
	if _, ok := body.Data["auditMode"]; !ok {
		t.Errorf("expected data.auditMode key, got keys: %v", body.Data)
	}
	if _, bad := body.Data["audit_mode"]; bad {
		t.Errorf("unexpected snake_case key audit_mode in response")
	}
	if v, ok := body.Data["auditMode"].(bool); !ok || v != true {
		t.Errorf("expected auditMode=true, got %v", body.Data["auditMode"])
	}
}

// TestGetConfig_ServiceError_Returns500 验证 service 出错时返回 500。
func TestGetConfig_ServiceError_Returns500(t *testing.T) {
	r := setupTest(&mockAppConfigService{
		getFn: func(ctx context.Context) (*AppConfigResponse, error) {
			return nil, context.DeadlineExceeded
		},
	})
	req := httptest.NewRequest("GET", "/app-config", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// setupPutTest 构造仅挂 PUT handler 的引擎（不含鉴权中间件，聚焦 handler 行为）。
func setupPutTest(mock *mockAppConfigService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := &AppConfigHandler{service: mock}
	r := gin.New()
	r.PUT("/app-config", h.UpdateConfig)
	return r
}

// TestUpdateConfig_Success 验证合法请求体更新成功并回读最新配置。
func TestUpdateConfig_Success(t *testing.T) {
	var gotAuditMode bool
	r := setupPutTest(&mockAppConfigService{
		updateFn: func(ctx context.Context, auditMode bool) error {
			gotAuditMode = auditMode
			return nil
		},
		getFn: func(ctx context.Context) (*AppConfigResponse, error) {
			return &AppConfigResponse{AuditMode: true}, nil
		},
	})
	req := httptest.NewRequest("PUT", "/app-config", strings.NewReader(`{"auditMode":true}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !gotAuditMode {
		t.Errorf("expected service to receive auditMode=true")
	}
}

// TestUpdateConfig_MissingAuditMode_Returns400 验证缺 auditMode 字段时返回 400。
func TestUpdateConfig_MissingAuditMode_Returns400(t *testing.T) {
	r := setupPutTest(&mockAppConfigService{
		updateFn: func(ctx context.Context, auditMode bool) error {
			t.Fatal("update should not be called when auditMode is missing")
			return nil
		},
	})
	req := httptest.NewRequest("PUT", "/app-config", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

// TestUpdateConfig_ServiceError_Returns500 验证 update service 出错时返回 500。
func TestUpdateConfig_ServiceError_Returns500(t *testing.T) {
	r := setupPutTest(&mockAppConfigService{
		updateFn: func(ctx context.Context, auditMode bool) error {
			return context.DeadlineExceeded
		},
	})
	req := httptest.NewRequest("PUT", "/app-config", strings.NewReader(`{"auditMode":false}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}
