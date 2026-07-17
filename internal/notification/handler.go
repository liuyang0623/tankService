package notification

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-service/pkg/response"
)

// notificationServiceIface abstracts NotificationService for handler injection/testability.
type notificationServiceIface interface {
	List(ctx context.Context, userID uint, page, limit int) (*PaginatedNotifications, error)
	MarkAllRead(ctx context.Context, userID uint) error
	UnreadSummary(ctx context.Context, userID uint) (*UnreadSummary, error)
}

// NotificationHandler handles HTTP requests for system notifications.
type NotificationHandler struct {
	service notificationServiceIface
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(service *NotificationService) *NotificationHandler {
	return &NotificationHandler{service: service}
}

func getUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := val.(uint)
	return uid, ok
}

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

// List godoc
// @Summary List my system notifications
// @Tags notification
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /notifications [get]
func (h *NotificationHandler) List(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	page, limit := parsePagination(c)
	result, err := h.service.List(c.Request.Context(), uid, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

// MarkRead godoc
// @Summary Mark all my notifications as read
// @Tags notification
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /notifications/read [post]
func (h *NotificationHandler) MarkRead(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	if err := h.service.MarkAllRead(c.Request.Context(), uid); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// UnreadCount godoc
// @Summary Get my unread notification count and latest summary
// @Tags notification
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /notifications/unread-count [get]
func (h *NotificationHandler) UnreadCount(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	result, err := h.service.UnreadSummary(c.Request.Context(), uid)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}
