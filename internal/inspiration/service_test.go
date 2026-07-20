package inspiration

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

// ===================== computeStreakFrom 纯函数测试 =====================

func TestComputeStreakFrom(t *testing.T) {
	base := time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)
	yesterday := base.AddDate(0, 0, -1)
	twoDaysAgo := base.AddDate(0, 0, -2)

	cases := []struct {
		name          string
		last          *time.Time
		currentStreak int
		want          int
	}{
		{"首次打卡 last=nil", nil, 0, 1},
		{"昨天已打卡今天连续+1", &yesterday, 3, 4},
		{"昨天已打卡但 streak=0 兜底", &yesterday, 0, 1},
		{"漏打一天重置为1", &twoDaysAgo, 5, 1},
		{"同日重复保持当前streak", &base, 3, 3},
		{"同日但 streak<1 兜底为1", &base, 0, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := computeStreakFrom(tc.last, base, tc.currentStreak)
			if got != tc.want {
				t.Errorf("computeStreakFrom(%v, base, %d) = %d, want %d",
					tc.last, tc.currentStreak, got, tc.want)
			}
		})
	}
}

func TestSameDayAndTruncate(t *testing.T) {
	morning := time.Date(2026, 7, 20, 8, 30, 0, 0, time.Local)
	night := time.Date(2026, 7, 20, 23, 59, 0, 0, time.Local)
	nextDay := time.Date(2026, 7, 21, 0, 1, 0, 0, time.Local)

	if !sameDay(morning, night) {
		t.Error("同一天不同时刻应判定为 sameDay")
	}
	if sameDay(night, nextDay) {
		t.Error("跨自然日不应为 sameDay")
	}
	if !truncateDay(morning).Equal(time.Date(2026, 7, 20, 0, 0, 0, 0, time.Local)) {
		t.Error("truncateDay 应截断到 00:00")
	}
}

// ===================== QA Handler 测试（mock service） =====================

type mockQAService struct {
	createQuestionFn func(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error)
	listQuestionsFn  func(ctx context.Context, page, limit int) (*PaginatedResult, error)
	getQuestionFn    func(ctx context.Context, id uint) (*QuestionResponse, error)
	createAnswerFn   func(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error)
}

func (m *mockQAService) CreateQuestion(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error) {
	return m.createQuestionFn(ctx, userID, input)
}
func (m *mockQAService) ListQuestions(ctx context.Context, page, limit int) (*PaginatedResult, error) {
	return m.listQuestionsFn(ctx, page, limit)
}
func (m *mockQAService) GetQuestion(ctx context.Context, id uint) (*QuestionResponse, error) {
	return m.getQuestionFn(ctx, id)
}
func (m *mockQAService) CreateAnswer(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error) {
	return m.createAnswerFn(ctx, questionID, userID, input)
}

func setupQA() (*gin.Engine, *mockQAService) {
	gin.SetMode(gin.TestMode)
	mock := &mockQAService{}
	h := &QAHandler{service: mock}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userID", uint(1)) })
	r.POST("/questions", h.CreateQuestion)
	r.GET("/questions", h.ListQuestions)
	r.GET("/questions/:id", h.GetQuestion)
	r.POST("/questions/:id/answers", h.CreateAnswer)
	return r, mock
}

func TestCreateQuestion_Success(t *testing.T) {
	r, mock := setupQA()
	mock.createQuestionFn = func(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error) {
		if userID != 1 {
			t.Errorf("expected userID=1, got %d", userID)
		}
		return &QuestionResponse{ID: 1, Title: input.Title, Content: input.Content, Answers: []AnswerResponse{}}, nil
	}
	body := `{"title":"如何坚持早起","content":"求方法"}`
	req := httptest.NewRequest("POST", "/questions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["title"] != "如何坚持早起" {
		t.Errorf("unexpected title: %v", data["title"])
	}
}

func TestCreateQuestion_EmptyTitle(t *testing.T) {
	r, _ := setupQA()
	body := `{"title":"","content":"x"}`
	req := httptest.NewRequest("POST", "/questions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListQuestions_Success(t *testing.T) {
	r, mock := setupQA()
	mock.listQuestionsFn = func(ctx context.Context, page, limit int) (*PaginatedResult, error) {
		return &PaginatedResult{
			Data: []QuestionListItem{{ID: 1, Title: "Q1", AnswerCount: 2}},
			Meta: PaginationMeta{Total: 1, Page: 1, Limit: 10, TotalPages: 1},
		}, nil
	}
	req := httptest.NewRequest("GET", "/questions?page=1&limit=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	items := data["data"].([]interface{})
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestGetQuestion_NotFound(t *testing.T) {
	r, mock := setupQA()
	mock.getQuestionFn = func(ctx context.Context, id uint) (*QuestionResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	req := httptest.NewRequest("GET", "/questions/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCreateAnswer_Success(t *testing.T) {
	r, mock := setupQA()
	mock.createAnswerFn = func(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error) {
		if questionID != 5 {
			t.Errorf("expected questionID=5, got %d", questionID)
		}
		return &AnswerResponse{ID: 10, AuthorID: userID, Content: input.Content}, nil
	}
	body := `{"content":"早睡就好"}`
	req := httptest.NewRequest("POST", "/questions/5/answers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateAnswer_EmptyContent(t *testing.T) {
	r, _ := setupQA()
	body := `{"content":""}`
	req := httptest.NewRequest("POST", "/questions/5/answers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateAnswer_QuestionNotFound(t *testing.T) {
	r, mock := setupQA()
	mock.createAnswerFn = func(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	body := `{"content":"x"}`
	req := httptest.NewRequest("POST", "/questions/999/answers", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestQA_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := &QAHandler{service: &mockQAService{}}
	r := gin.New() // 无 JWT 中间件
	r.POST("/questions", h.CreateQuestion)
	body := `{"title":"x"}`
	req := httptest.NewRequest("POST", "/questions", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// ===================== Sport Handler 测试（mock service） =====================

type mockSportService struct {
	createGoalFn func(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error)
	listGoalsFn  func(ctx context.Context, userID uint) ([]SportGoalResponse, error)
	updateGoalFn func(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error)
	checkinFn    func(ctx context.Context, id, userID uint) (*CheckinResponse, error)
}

func (m *mockSportService) CreateGoal(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error) {
	return m.createGoalFn(ctx, userID, input)
}
func (m *mockSportService) ListGoals(ctx context.Context, userID uint) ([]SportGoalResponse, error) {
	return m.listGoalsFn(ctx, userID)
}
func (m *mockSportService) UpdateGoal(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error) {
	return m.updateGoalFn(ctx, id, userID, input)
}
func (m *mockSportService) Checkin(ctx context.Context, id, userID uint) (*CheckinResponse, error) {
	return m.checkinFn(ctx, id, userID)
}

func setupSport() (*gin.Engine, *mockSportService) {
	gin.SetMode(gin.TestMode)
	mock := &mockSportService{}
	h := &SportHandler{service: mock}
	r := gin.New()
	r.Use(func(c *gin.Context) { c.Set("userID", uint(1)) })
	r.GET("/sport-goals", h.ListGoals)
	r.POST("/sport-goals", h.CreateGoal)
	r.PATCH("/sport-goals/:id", h.UpdateGoal)
	r.POST("/sport-goals/:id/checkin", h.Checkin)
	return r, mock
}

func TestCreateGoal_Success(t *testing.T) {
	r, mock := setupSport()
	mock.createGoalFn = func(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error) {
		return &SportGoalResponse{ID: 1, Name: input.Name, TargetDays: input.TargetDays}, nil
	}
	body := `{"name":"每天跑步","type":"running","targetDays":30}`
	req := httptest.NewRequest("POST", "/sport-goals", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateGoal_EmptyName(t *testing.T) {
	r, _ := setupSport()
	body := `{"name":""}`
	req := httptest.NewRequest("POST", "/sport-goals", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestListGoals_Success(t *testing.T) {
	r, mock := setupSport()
	mock.listGoalsFn = func(ctx context.Context, userID uint) ([]SportGoalResponse, error) {
		return []SportGoalResponse{{ID: 1, Name: "跑步", Streak: 3, TotalDays: 5, CheckedInToday: true}}, nil
	}
	req := httptest.NewRequest("GET", "/sport-goals", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUpdateGoal_NotFound(t *testing.T) {
	r, mock := setupSport()
	mock.updateGoalFn = func(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	body := `{"name":"改名"}`
	req := httptest.NewRequest("PATCH", "/sport-goals/999", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCheckin_Success(t *testing.T) {
	r, mock := setupSport()
	mock.checkinFn = func(ctx context.Context, id, userID uint) (*CheckinResponse, error) {
		return &CheckinResponse{GoalID: id, Streak: 4, TotalDays: 6, CheckedInToday: true, Awarded: true}, nil
	}
	req := httptest.NewRequest("POST", "/sport-goals/1/checkin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["awarded"] != true {
		t.Errorf("expected awarded=true, got %v", data["awarded"])
	}
}

func TestCheckin_GoalNotFound(t *testing.T) {
	r, mock := setupSport()
	mock.checkinFn = func(ctx context.Context, id, userID uint) (*CheckinResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	req := httptest.NewRequest("POST", "/sport-goals/999/checkin", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
