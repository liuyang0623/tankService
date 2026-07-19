package users

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// userServiceIface abstracts UserService for handler injection and testing.
type userServiceIface interface {
	GetProfile(ctx context.Context, userID uint) (*User, error)
	FindOne(ctx context.Context, userID uint) (*User, error)
	UpdateProfile(ctx context.Context, userID uint, updates map[string]interface{}) (*User, error)
	IncrSubscribeFollowQuota(ctx context.Context, userID uint) error
}

// followStatsIface provides follow counts and follow state without importing the
// follow package (avoids a users<->follow import cycle). Injected from main.
type followStatsIface interface {
	CountFollowers(ctx context.Context, userID uint) (int64, error)
	CountFollowing(ctx context.Context, userID uint) (int64, error)
	IsFollowing(ctx context.Context, currentUserID, targetID uint) (bool, error)
}

// UserDetailResponse is the API DTO for GET /users/:id, extending the user with
// follow counts and the current user's follow state.
type UserDetailResponse struct {
	*User
	FollowerCount  int64 `json:"followerCount"`
	FollowingCount int64 `json:"followingCount"`
	IsFollowing    bool  `json:"isFollowing"`
}

// UserHandler handles HTTP requests for users.
type UserHandler struct {
	service     userServiceIface
	followStats followStatsIface // optional; when nil, counts default to 0
}

// NewUserHandler creates a new UserHandler with the given service.
func NewUserHandler(service *UserService) *UserHandler {
	return &UserHandler{service: service}
}

// SetFollowStats injects the follow-stats provider (called from main after both
// services are constructed).
func (h *UserHandler) SetFollowStats(fs followStatsIface) {
	h.followStats = fs
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

// GetProfile godoc
// @Summary Get current user profile
// @Tags users
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /users/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	user, err := h.service.GetProfile(c.Request.Context(), uid)
	if err != nil {
		// token 有效但用户在库中不存在（跨环境旧 token、账号已删除等）视为登录态失效，
		// 返回 401 触发前端重新登录，而非 404。
		if err == gorm.ErrRecordNotFound {
			response.Unauthorized(c, "login expired")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, user)
}

// SubscribeFollow godoc
// @Summary 上报关注订阅授权（累加可推送配额）
// @Tags users
// @Security Bearer
// @Success 200 {object} map[string]interface{}
// @Router /users/subscribe/follow [post]
func (h *UserHandler) SubscribeFollow(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}
	if err := h.service.IncrSubscribeFollowQuota(c.Request.Context(), uid); err != nil {
		response.InternalError(c, "server error")
		return
	}
	response.Success(c, gin.H{"ok": true})
}

// UpdateProfile godoc
// @Summary Update current user profile
// @Tags users
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body map[string]interface{} true "Profile updates"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/profile [patch]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var rawUpdates map[string]interface{}
	if err := c.ShouldBindJSON(&rawUpdates); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	// Filter request body to the whitelist of allowed profile fields
	// before passing it to the service layer.
	updates := make(map[string]interface{}, len(rawUpdates))
	for k, v := range rawUpdates {
		if allowedUpdateFields[k] {
			updates[k] = v
		}
	}

	user, err := h.service.UpdateProfile(c.Request.Context(), uid, updates)
	if err != nil {
		// 当前登录用户在库中不存在，同 GetProfile 视为登录态失效，返回 401。
		if err == gorm.ErrRecordNotFound {
			response.Unauthorized(c, "login expired")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, user)
}

// GetUser godoc
// @Summary Get user by ID
// @Tags users
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.service.FindOne(c.Request.Context(), uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "user not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	resp := UserDetailResponse{User: user}
	if h.followStats != nil {
		ctx := c.Request.Context()
		if fc, err := h.followStats.CountFollowers(ctx, uint(id)); err == nil {
			resp.FollowerCount = fc
		}
		if gc, err := h.followStats.CountFollowing(ctx, uint(id)); err == nil {
			resp.FollowingCount = gc
		}
		// current logged-in user (optional auth); 0 when not logged in
		if currentID, ok := getUserID(c); ok {
			if isF, err := h.followStats.IsFollowing(ctx, currentID, uint(id)); err == nil {
				resp.IsFollowing = isF
			}
		}
	}

	response.Success(c, resp)
}
