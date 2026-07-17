package response_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

func setupRouter(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)
	r.GET("/test", handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	c.Request = req
	r.ServeHTTP(w, req)
	return w
}

// --- Success ---

func TestSuccess_StatusCode(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Success(c, gin.H{"id": 1})
	})
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestSuccess_Body(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Success(c, gin.H{"id": 1})
	})

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["message"] != "success" {
		t.Errorf("expected message 'success', got %v", body["message"])
	}
	if body["code"] != float64(http.StatusOK) {
		t.Errorf("expected code 200, got %v", body["code"])
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected data to be a map, got %T: %v", body["data"], body["data"])
	}
	id, ok := data["id"]
	if !ok {
		t.Error("expected data.id to exist")
	}
	// JSON numbers decode as float64
	if id != float64(1) {
		t.Errorf("expected data.id == 1, got %v", id)
	}
}

func TestSuccess_NilData(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Success(c, nil)
	})
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "success" {
		t.Errorf("expected message 'success', got %v", body["message"])
	}
	if body["code"] != float64(http.StatusOK) {
		t.Errorf("expected code 200, got %v", body["code"])
	}
}

// --- Error ---

func TestError_StatusCode(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Error(c, http.StatusBadRequest, "invalid input")
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestError_Body(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Error(c, http.StatusBadRequest, "invalid input")
	})

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "invalid input" {
		t.Errorf("expected message 'invalid input', got %v", body["message"])
	}
	if body["code"] != float64(http.StatusBadRequest) {
		t.Errorf("expected code %d, got %v", http.StatusBadRequest, body["code"])
	}
	if body["data"] != nil {
		t.Errorf("expected data to be null, got %v", body["data"])
	}
}

func TestError_InternalServerError(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Error(c, http.StatusInternalServerError, "something went wrong")
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// --- Convenience helpers ---

func TestBadRequest(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.BadRequest(c, "bad request error")
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["message"] != "bad request error" {
		t.Errorf("expected message 'bad request error', got %v", body["message"])
	}
	if body["code"] != float64(http.StatusBadRequest) {
		t.Errorf("expected code 400, got %v", body["code"])
	}
}

func TestUnauthorized(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.Unauthorized(c, "unauthorized")
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["code"] != float64(http.StatusUnauthorized) {
		t.Errorf("expected code 401, got %v", body["code"])
	}
}

func TestInternalError(t *testing.T) {
	w := setupRouter(func(c *gin.Context) {
		response.InternalError(c, "internal error")
	})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if body["code"] != float64(http.StatusInternalServerError) {
		t.Errorf("expected code 500, got %v", body["code"])
	}
}
