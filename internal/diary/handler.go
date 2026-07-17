package diary

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// diaryServiceIface abstracts DiaryService for handler injection and testing.
type diaryServiceIface interface {
	Create(ctx context.Context, userID uint, input CreateDiaryInput) (*DiaryResponse, error)
	FindMine(ctx context.Context, userID uint, notebookID uint, page, limit int) (*PaginatedResult, error)
	FindOne(ctx context.Context, id, userID uint) (*DiaryResponse, error)
	Update(ctx context.Context, id, userID uint, input UpdateDiaryInput) (*DiaryResponse, error)
	Remove(ctx context.Context, id, userID uint) error
}

// DiaryHandler handles HTTP requests for diary entries.
type DiaryHandler struct {
	service diaryServiceIface
}

// NewDiaryHandler creates a new DiaryHandler with the given service.
func NewDiaryHandler(service *DiaryService) *DiaryHandler {
	return &DiaryHandler{service: service}
}

// getUserID retrieves the userID injected by JWTMiddleware from gin context.
func getUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := val.(uint)
	return uid, ok
}

// Create godoc
// @Summary Create a new diary entry
// @Tags diaries
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body CreateDiaryInput true "Diary data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /diaries [post]
func (h *DiaryHandler) Create(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreateDiaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if input.Title == "" {
		response.BadRequest(c, "title is required")
		return
	}
	if input.Content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	diary, err := h.service.Create(c.Request.Context(), uid, input)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, diary)
}

// FindMine godoc
// @Summary Get current user's diary list (timeline)
// @Tags diaries
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /diaries [get]
func (h *DiaryHandler) FindMine(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit := parsePagination(c)

	var notebookID uint
	if v, err := strconv.ParseUint(c.Query("notebookId"), 10, 32); err == nil {
		notebookID = uint(v)
	}

	result, err := h.service.FindMine(c.Request.Context(), uid, notebookID, page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// FindOne godoc
// @Summary Get a diary entry by ID
// @Tags diaries
// @Security Bearer
// @Param id path int true "Diary ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /diaries/{id} [get]
func (h *DiaryHandler) FindOne(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid diary id")
		return
	}

	diary, err := h.service.FindOne(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "diary not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, diary)
}

// Update godoc
// @Summary Update a diary entry
// @Tags diaries
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Diary ID"
// @Param body body UpdateDiaryInput true "Update fields"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /diaries/{id} [patch]
func (h *DiaryHandler) Update(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid diary id")
		return
	}

	var input UpdateDiaryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	diary, err := h.service.Update(c.Request.Context(), uint(id), uid, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "diary not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, diary)
}

// Remove godoc
// @Summary Delete a diary entry
// @Tags diaries
// @Security Bearer
// @Param id path int true "Diary ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /diaries/{id} [delete]
func (h *DiaryHandler) Remove(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid diary id")
		return
	}

	err = h.service.Remove(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "diary not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, gin.H{"message": "deleted"})
}

// parsePagination extracts page and limit from query params.
func parsePagination(c *gin.Context) (int, int) {
	pageStr := c.Query("page")
	limitStr := c.Query("limit")

	page := 1
	limit := 10

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	return page, limit
}