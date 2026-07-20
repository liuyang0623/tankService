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
	createQuestionFn   func(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error)
	listQuestionsFn    func(ctx context.Context, page, limit int) (*PaginatedResult, error)
	getQuestionFn      func(ctx context.Context, id, userID uint) (*QuestionResponse, error)
	createAnswerFn     func(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error)
	toggleAnswerLikeFn func(ctx context.Context, questionID, answerID, userID uint) (*AnswerLikeResponse, error)
	acceptAnswerFn     func(ctx context.Context, questionID, answerID, userID uint) error
}

func (m *mockQAService) CreateQuestion(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error) {
	return m.createQuestionFn(ctx, userID, input)
}
func (m *mockQAService) ListQuestions(ctx context.Context, page, limit int) (*PaginatedResult, error) {
	return m.listQuestionsFn(ctx, page, limit)
}
func (m *mockQAService) GetQuestion(ctx context.Context, id, userID uint) (*QuestionResponse, error) {
	return m.getQuestionFn(ctx, id, userID)
}
func (m *mockQAService) CreateAnswer(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error) {
	return m.createAnswerFn(ctx, questionID, userID, input)
}
func (m *mockQAService) ToggleAnswerLike(ctx context.Context, questionID, answerID, userID uint) (*AnswerLikeResponse, error) {
	return m.toggleAnswerLikeFn(ctx, questionID, answerID, userID)
}
func (m *mockQAService) AcceptAnswer(ctx context.Context, questionID, answerID, userID uint) error {
	return m.acceptAnswerFn(ctx, questionID, answerID, userID)
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
	r.POST("/questions/:id/answers/:aid/like", h.ToggleAnswerLike)
	r.POST("/questions/:id/accept/:aid", h.AcceptAnswer)
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
	mock.getQuestionFn = func(ctx context.Context, id, userID uint) (*QuestionResponse, error) {
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
	createGoalFn       func(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error)
	listGoalsFn        func(ctx context.Context, userID uint) ([]SportGoalResponse, error)
	updateGoalFn       func(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error)
	checkinFn          func(ctx context.Context, id, userID uint) (*CheckinResponse, error)
	listMonthRecordsFn func(ctx context.Context, goalID, userID uint, year, month int) (*MonthRecordsResponse, error)
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
func (m *mockSportService) ListMonthRecords(ctx context.Context, goalID, userID uint, year, month int) (*MonthRecordsResponse, error) {
	return m.listMonthRecordsFn(ctx, goalID, userID, year, month)
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
	r.GET("/sport-goals/:id/records", h.ListMonthRecords)
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

// ===================== 排序与最佳标记纯函数测试 =====================

func TestMarkBestAnswer(t *testing.T) {
	t0 := time.Date(2026, 7, 20, 10, 0, 0, 0, time.Local)
	t1 := t0.Add(time.Hour)
	t2 := t0.Add(2 * time.Hour)

	t.Run("赞数最高标记为最佳", func(t *testing.T) {
		as := []AnswerResponse{
			{ID: 1, LikeCount: 2, CreatedAt: t0},
			{ID: 2, LikeCount: 5, CreatedAt: t1},
			{ID: 3, LikeCount: 3, CreatedAt: t2},
		}
		markBestAnswer(as)
		if as[1].ID != 2 || !as[1].IsBest {
			t.Errorf("expected answer 2 be best")
		}
		if as[0].IsBest || as[2].IsBest {
			t.Errorf("only one best expected")
		}
	})

	t.Run("全部0赞则无最佳", func(t *testing.T) {
		as := []AnswerResponse{{ID: 1, LikeCount: 0, CreatedAt: t0}, {ID: 2, LikeCount: 0, CreatedAt: t1}}
		markBestAnswer(as)
		for _, a := range as {
			if a.IsBest {
				t.Errorf("no best expected when all zero likes")
			}
		}
	})

	t.Run("赞数并列取最早", func(t *testing.T) {
		as := []AnswerResponse{
			{ID: 1, LikeCount: 4, CreatedAt: t2},
			{ID: 2, LikeCount: 4, CreatedAt: t0},
		}
		markBestAnswer(as)
		if !as[1].IsBest || as[0].IsBest {
			t.Errorf("earliest among ties should be best")
		}
	})
}

func TestSortAnswers(t *testing.T) {
	t0 := time.Date(2026, 7, 20, 10, 0, 0, 0, time.Local)
	t1 := t0.Add(time.Hour)
	t2 := t0.Add(2 * time.Hour)

	as := []AnswerResponse{
		{ID: 1, LikeCount: 5, CreatedAt: t0},                   // 高赞但未采纳
		{ID: 2, LikeCount: 1, CreatedAt: t1, IsAccepted: true}, // 采纳 → 置顶
		{ID: 3, LikeCount: 3, CreatedAt: t2},
		{ID: 4, LikeCount: 3, CreatedAt: t0}, // 与3同赞但更早
	}
	sortAnswers(as)

	wantOrder := []uint{2, 1, 4, 3} // 采纳 → 赞降序 → 同赞早答优先
	for i, want := range wantOrder {
		if as[i].ID != want {
			t.Errorf("position %d: got id %d, want %d", i, as[i].ID, want)
		}
	}
}

// ===================== 新增 Handler 测试 =====================

func TestToggleAnswerLike_Success(t *testing.T) {
	r, mock := setupQA()
	mock.toggleAnswerLikeFn = func(ctx context.Context, questionID, answerID, userID uint) (*AnswerLikeResponse, error) {
		if questionID != 5 || answerID != 10 {
			t.Errorf("unexpected ids q=%d a=%d", questionID, answerID)
		}
		return &AnswerLikeResponse{AnswerID: answerID, Liked: true, LikeCount: 3}, nil
	}
	req := httptest.NewRequest("POST", "/questions/5/answers/10/like", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	if data["liked"] != true || data["likeCount"].(float64) != 3 {
		t.Errorf("unexpected like response: %v", data)
	}
}

func TestToggleAnswerLike_NotFound(t *testing.T) {
	r, mock := setupQA()
	mock.toggleAnswerLikeFn = func(ctx context.Context, questionID, answerID, userID uint) (*AnswerLikeResponse, error) {
		return nil, gorm.ErrRecordNotFound
	}
	req := httptest.NewRequest("POST", "/questions/5/answers/999/like", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestAcceptAnswer_Success(t *testing.T) {
	r, mock := setupQA()
	mock.acceptAnswerFn = func(ctx context.Context, questionID, answerID, userID uint) error {
		return nil
	}
	req := httptest.NewRequest("POST", "/questions/5/accept/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAcceptAnswer_Forbidden(t *testing.T) {
	r, mock := setupQA()
	mock.acceptAnswerFn = func(ctx context.Context, questionID, answerID, userID uint) error {
		return errNotAuthor
	}
	req := httptest.NewRequest("POST", "/questions/5/accept/10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestListMonthRecords_Success(t *testing.T) {
	r, mock := setupSport()
	mock.listMonthRecordsFn = func(ctx context.Context, goalID, userID uint, year, month int) (*MonthRecordsResponse, error) {
		if year != 2026 || month != 7 {
			t.Errorf("unexpected year/month %d/%d", year, month)
		}
		return &MonthRecordsResponse{Year: year, Month: month, Dates: []string{"2026-07-19", "2026-07-20"}}, nil
	}
	req := httptest.NewRequest("GET", "/sport-goals/1/records?month=2026-07", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	dates := data["dates"].([]interface{})
	if len(dates) != 2 {
		t.Errorf("expected 2 dates, got %d", len(dates))
	}
}

func TestListMonthRecords_InvalidMonth(t *testing.T) {
	r, _ := setupSport()
	req := httptest.NewRequest("GET", "/sport-goals/1/records?month=2026-13-x", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
