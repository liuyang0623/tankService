package notebook

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// notebookServiceIface 抽象 NotebookService 供 handler 注入与测试。
type notebookServiceIface interface {
	Create(ctx context.Context, userID uint, input CreateNotebookInput) (*NotebookResponse, error)
	FindMine(ctx context.Context, userID uint) ([]NotebookResponse, error)
	Update(ctx context.Context, id, userID uint, input UpdateNotebookInput) (*NotebookResponse, error)
	Remove(ctx context.Context, id, userID uint) error
}

// NotebookHandler 处理日记本 HTTP 请求。
type NotebookHandler struct {
	service notebookServiceIface
}

// NewNotebookHandler 构造。
func NewNotebookHandler(service *NotebookService) *NotebookHandler {
	return &NotebookHandler{service: service}
}

// getUserID 从 gin context 取 JWTMiddleware 注入的 userID。
func getUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := val.(uint)
	return uid, ok
}

// Create 创建日记本。
func (h *NotebookHandler) Create(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	var input CreateNotebookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	if input.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}
	nb, err := h.service.Create(c.Request.Context(), uid, input)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, nb)
}

// FindMine 返回当前用户的日记本列表。
func (h *NotebookHandler) FindMine(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	list, err := h.service.FindMine(c.Request.Context(), uid)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, list)
}

// Update 更新日记本。
func (h *NotebookHandler) Update(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid notebook id")
		return
	}
	var input UpdateNotebookInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}
	nb, err := h.service.Update(c.Request.Context(), uint(id), uid, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "notebook not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, nb)
}

// Remove 删除日记本。
func (h *NotebookHandler) Remove(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid notebook id")
		return
	}
	if err := h.service.Remove(c.Request.Context(), uint(id), uid); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "notebook not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, gin.H{"message": "deleted"})
}
