package message

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"go-service/pkg/middleware"
	"go-service/pkg/response"
)

// messageServiceIface abstracts MessageService for handler injection/testability.
type messageServiceIface interface {
	CreateMessage(ctx context.Context, senderID, toUserID uint, msgType, content string) (*Message, error)
	GetConversations(ctx context.Context, userID uint, page, limit int) (*PaginatedConversations, error)
	GetMessages(ctx context.Context, conversationID, userID uint, page, limit int) (*PaginatedMessages, error)
	MarkRead(ctx context.Context, conversationID, userID uint) error
	FindConversationByUsers(ctx context.Context, userID, otherUserID uint) (uint, error)
}

// MessageHandler handles HTTP requests for messaging.
type MessageHandler struct {
	service  messageServiceIface
	hub      *Hub
	jwtSecret string
	upgrader websocket.Upgrader
}

// NewMessageHandler creates a new MessageHandler.
func NewMessageHandler(service *MessageService, hub *Hub, jwtSecret string) *MessageHandler {
	return &MessageHandler{
		service:  service,
		hub:      hub,
		jwtSecret: jwtSecret,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in dev
			},
		},
	}
}

// getUserID retrieves the userID injected by JWTMiddleware.
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

func parseConversationID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// SendMessage godoc
// @Summary Send a message to another user
// @Tags message
// @Security Bearer
// @Param body body SendMessageRequest true "Message body"
// @Success 200 {object} map[string]interface{}
// @Router /messages [post]
func (h *MessageHandler) SendMessage(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: "+err.Error())
		return
	}

	if req.Type == "" {
		req.Type = "text"
	}

	msg, err := h.service.CreateMessage(c.Request.Context(), uid, req.ToUserID, req.Type, req.Content)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "user not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}

	// Push to receiver's WebSocket if online
	wsMsg := map[string]interface{}{
		"type": "new_message",
		"data": msg.toMessageItem(),
	}
	h.pushJSON(req.ToUserID, wsMsg)

	response.Success(c, msg.toMessageItem())
}

// ListConversations godoc
// @Summary List the current user's conversations
// @Tags message
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /conversations [get]
func (h *MessageHandler) ListConversations(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	// 若带 withUser 参数，返回与该用户的会话 id（0 表示无会话），供主页私信入口定位历史
	if withUser := c.Query("withUser"); withUser != "" {
		otherID, err := strconv.ParseUint(withUser, 10, 32)
		if err != nil {
			response.BadRequest(c, "invalid withUser id")
			return
		}
		convID, err := h.service.FindConversationByUsers(c.Request.Context(), uid, uint(otherID))
		if err != nil {
			response.InternalError(c, err.Error())
			return
		}
		response.Success(c, gin.H{"conversationId": convID})
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.GetConversations(c.Request.Context(), uid, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// GetMessages godoc
// @Summary Get paginated messages for a conversation
// @Tags message
// @Security Bearer
// @Param id path int true "Conversation ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /conversations/{id}/messages [get]
func (h *MessageHandler) GetMessages(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	convID, ok := parseConversationID(c)
	if !ok {
		response.BadRequest(c, "invalid conversation id")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.GetMessages(c.Request.Context(), convID, uid, page, limit)
	if err != nil {
		if err.Error() == "not a participant of this conversation" {
			response.Error(c, http.StatusForbidden, err.Error())
			return
		}
		response.InternalError(c, err.Error())
		return
	}
	response.Success(c, result)
}

// MarkRead godoc
// @Summary Mark messages as read in a conversation
// @Tags message
// @Security Bearer
// @Param id path int true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Router /conversations/{id}/read [post]
func (h *MessageHandler) MarkRead(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	convID, ok := parseConversationID(c)
	if !ok {
		response.BadRequest(c, "invalid conversation id")
		return
	}

	if err := h.service.MarkRead(c.Request.Context(), convID, uid); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "conversation not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"success": true})
}

// HandleWS upgrades an HTTP connection to WebSocket.
// The JWT token must be provided as a query parameter: /ws?token=<jwt>
func (h *MessageHandler) HandleWS(c *gin.Context) {
	tokenString := c.Query("token")
	if tokenString == "" {
		response.BadRequest(c, "token query parameter required")
		return
	}

	claims, err := middleware.ParseToken(tokenString, h.jwtSecret)
	if err != nil {
		response.Unauthorized(c, "invalid or expired token")
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	h.hub.ServeWS(conn, claims.UserID)
}

// pushJSON marshals v as JSON and sends it to the user's WebSocket connections.
func (h *MessageHandler) pushJSON(userID uint, v interface{}) {
	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("[ws] marshal error: %v", err)
		return
	}
	h.hub.SendToUser(userID, data)
}
