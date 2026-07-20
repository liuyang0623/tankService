package inspiration

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// getUserID 从 gin context 读取 JWTMiddleware 注入的 userID。
func getUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := val.(uint)
	return uid, ok
}

// parsePagination 从 query 解析 page 与 limit，缺省 page=1、limit=10。
func parsePagination(c *gin.Context) (int, int) {
	page := 1
	limit := 10
	if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	return page, limit
}

// ===================== 解惑问答 Handler =====================

// qaServiceIface 抽象 QAService，便于 handler 注入与测试。
type qaServiceIface interface {
	CreateQuestion(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error)
	ListQuestions(ctx context.Context, page, limit int) (*PaginatedResult, error)
	GetQuestion(ctx context.Context, id, userID uint) (*QuestionResponse, error)
	CreateAnswer(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error)
	ToggleAnswerLike(ctx context.Context, questionID, answerID, userID uint) (*AnswerLikeResponse, error)
	AcceptAnswer(ctx context.Context, questionID, answerID, userID uint) error
}

// QAHandler 处理解惑问答的 HTTP 请求。
type QAHandler struct {
	service qaServiceIface
}

// NewQAHandler 创建 QAHandler。
func NewQAHandler(service *QAService) *QAHandler {
	return &QAHandler{service: service}
}

// CreateQuestion 提问。
func (h *QAHandler) CreateQuestion(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreateQuestionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if input.Title == "" {
		response.BadRequest(c, "title is required")
		return
	}

	q, err := h.service.CreateQuestion(c.Request.Context(), uid, input)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, q)
}

// ListQuestions 获取全站问题列表（分页）。
func (h *QAHandler) ListQuestions(c *gin.Context) {
	if _, ok := getUserID(c); !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit := parsePagination(c)
	result, err := h.service.ListQuestions(c.Request.Context(), page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, result)
}

// GetQuestion 获取问题详情（含回答）。
func (h *QAHandler) GetQuestion(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid question id")
		return
	}

	q, err := h.service.GetQuestion(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "question not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, q)
}

// CreateAnswer 回答问题。
func (h *QAHandler) CreateAnswer(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid question id")
		return
	}

	var input CreateAnswerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if input.Content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	a, err := h.service.CreateAnswer(c.Request.Context(), uint(id), uid, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "question not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, a)
}

// ToggleAnswerLike 点赞/取消点赞某回答。
func (h *QAHandler) ToggleAnswerLike(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	qid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid question id")
		return
	}
	aid, err := strconv.ParseUint(c.Param("aid"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid answer id")
		return
	}

	res, err := h.service.ToggleAnswerLike(c.Request.Context(), uint(qid), uint(aid), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "answer not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, res)
}

// AcceptAnswer 由提问者采纳某回答。
func (h *QAHandler) AcceptAnswer(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	qid, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid question id")
		return
	}
	aid, err := strconv.ParseUint(c.Param("aid"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid answer id")
		return
	}

	err = h.service.AcceptAnswer(c.Request.Context(), uint(qid), uint(aid), uid)
	if err != nil {
		switch {
		case errors.Is(err, errNotAuthor):
			response.Error(c, http.StatusForbidden, "only the question author can accept answers")
		case err == gorm.ErrRecordNotFound:
			response.Error(c, http.StatusNotFound, "question or answer not found")
		default:
			response.InternalError(c, "server error")
		}
		return
	}
	response.Success(c, gin.H{"questionId": uint(qid), "acceptedAnswerId": uint(aid)})
}

// ===================== 运动计划 Handler =====================

// sportServiceIface 抽象 SportService，便于 handler 注入与测试。
type sportServiceIface interface {
	CreateGoal(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error)
	ListGoals(ctx context.Context, userID uint) ([]SportGoalResponse, error)
	UpdateGoal(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error)
	Checkin(ctx context.Context, id, userID uint) (*CheckinResponse, error)
	ListMonthRecords(ctx context.Context, goalID, userID uint, year, month int) (*MonthRecordsResponse, error)
}

// SportHandler 处理运动计划的 HTTP 请求。
type SportHandler struct {
	service sportServiceIface
}

// NewSportHandler 创建 SportHandler。
func NewSportHandler(service *SportService) *SportHandler {
	return &SportHandler{service: service}
}

// ListGoals 获取当前用户的运动目标列表。
func (h *SportHandler) ListGoals(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	goals, err := h.service.ListGoals(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, goals)
}

// CreateGoal 创建运动目标。
func (h *SportHandler) CreateGoal(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreateSportGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if input.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	goal, err := h.service.CreateGoal(c.Request.Context(), uid, input)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, goal)
}

// UpdateGoal 更新运动目标。
func (h *SportHandler) UpdateGoal(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid goal id")
		return
	}

	var input UpdateSportGoalInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	goal, err := h.service.UpdateGoal(c.Request.Context(), uint(id), uid, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "goal not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, goal)
}

// Checkin 对运动目标打卡。
func (h *SportHandler) Checkin(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid goal id")
		return
	}

	result, err := h.service.Checkin(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "goal not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, result)
}

// ListMonthRecords 返回某目标指定月份的打卡日期列表。
// query ?month=YYYY-MM，缺省当月。
func (h *SportHandler) ListMonthRecords(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid goal id")
		return
	}

	now := time.Now()
	year, month := now.Year(), int(now.Month())
	if m := c.Query("month"); m != "" {
		parsed, perr := time.Parse("2006-01", m)
		if perr != nil {
			response.BadRequest(c, "invalid month format, expect YYYY-MM")
			return
		}
		year, month = parsed.Year(), int(parsed.Month())
	}

	res, err := h.service.ListMonthRecords(c.Request.Context(), uint(id), uid, year, month)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "goal not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, res)
}
