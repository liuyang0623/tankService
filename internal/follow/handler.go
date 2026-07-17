package follow

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// followServiceIface abstracts FollowService for handler injection/testability.
type followServiceIface interface {
	ToggleFollow(ctx context.Context, followerID, targetID uint) (bool, error)
	ListFollowers(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error)
	ListFollowing(ctx context.Context, userID, currentUserID uint, page, limit int) (*PaginatedUsers, error)
}

// FollowHandler handles HTTP requests for follow relationships.
type FollowHandler struct {
	service followServiceIface
}

// NewFollowHandler creates a new FollowHandler.
func NewFollowHandler(service *FollowService) *FollowHandler {
	return &FollowHandler{service: service}
}

// getUserID retrieves the userID injected by JWTMiddleware (required auth).
func getUserID(c *gin.Context) (uint, bool) {
	val, ok := c.Get("userID")
	if !ok {
		return 0, false
	}
	uid, ok := val.(uint)
	return uid, ok
}

// getOptionalUserID retrieves userID if present (OptionalJWTMiddleware); 0 if absent.
func getOptionalUserID(c *gin.Context) uint {
	uid, ok := getUserID(c)
	if !ok {
		return 0
	}
	return uid
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

func parseUserID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, false
	}
	return uint(id), true
}

// ToggleFollow godoc
// @Summary Follow or unfollow a user
// @Tags follow
// @Security Bearer
// @Param id path int true "Target user ID"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id}/follow [post]
func (h *FollowHandler) ToggleFollow(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	targetID, ok := parseUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user id")
		return
	}

	following, err := h.service.ToggleFollow(c.Request.Context(), uid, targetID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, "user not found")
			return
		}
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, gin.H{"following": following})
}

// ListFollowers godoc
// @Summary List followers of a user
// @Tags follow
// @Param id path int true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id}/followers [get]
func (h *FollowHandler) ListFollowers(c *gin.Context) {
	targetID, ok := parseUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user id")
		return
	}
	page, limit := parsePagination(c)
	currentUserID := getOptionalUserID(c)

	result, err := h.service.ListFollowers(c.Request.Context(), targetID, currentUserID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

// ListFollowing godoc
// @Summary List users that a user follows
// @Tags follow
// @Param id path int true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Router /users/{id}/following [get]
func (h *FollowHandler) ListFollowing(c *gin.Context) {
	targetID, ok := parseUserID(c)
	if !ok {
		response.BadRequest(c, "invalid user id")
		return
	}
	page, limit := parsePagination(c)
	currentUserID := getOptionalUserID(c)

	result, err := h.service.ListFollowing(c.Request.Context(), targetID, currentUserID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}
