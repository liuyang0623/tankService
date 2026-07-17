package interactions

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// interactionServiceIface abstracts InteractionService for handler injection.
type interactionServiceIface interface {
	LikePost(ctx context.Context, userID, postID uint) (bool, error)
	FavoritePost(ctx context.Context, userID, postID uint) (bool, error)
	GetUserFavorites(ctx context.Context, userID uint, page, limit int) (*PaginatedFavorites, error)
	CreateComment(ctx context.Context, userID, postID uint, content string, parentID *uint) (*CommentResponse, error)
	GetPostComments(ctx context.Context, postID uint, page, limit int) (*PaginatedComments, error)
	DeleteComment(ctx context.Context, userID, commentID uint) error
}

// InteractionHandler handles HTTP requests for interactions.
type InteractionHandler struct {
	service interactionServiceIface
}

// NewInteractionHandler creates a new InteractionHandler.
func NewInteractionHandler(service *InteractionService) *InteractionHandler {
	return &InteractionHandler{service: service}
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

// LikePost godoc
// @Summary Like or unlike a post
// @Tags interactions
// @Security Bearer
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id}/like [post]
func (h *InteractionHandler) LikePost(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	liked, err := h.service.LikePost(c.Request.Context(), uid, uint(id))
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, gin.H{"liked": liked})
}

// FavoritePost godoc
// @Summary Favorite or unfavorite a post
// @Tags interactions
// @Security Bearer
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id}/favorite [post]
func (h *InteractionHandler) FavoritePost(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	favorited, err := h.service.FavoritePost(c.Request.Context(), uid, uint(id))
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, gin.H{"favorited": favorited})
}

// GetUserFavorites godoc
// @Summary Get current user's favorited posts
// @Tags interactions
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/me/favorites [get]
func (h *InteractionHandler) GetUserFavorites(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.GetUserFavorites(c.Request.Context(), uid, page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// CreateCommentRequest is the JSON body for creating a comment.
type CreateCommentRequest struct {
	PostID   uint   `json:"postId"`
	Content  string `json:"content"`
	ParentID *uint  `json:"parentId,omitempty"`
}

// CreateComment godoc
// @Summary Create a comment on a post
// @Tags interactions
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body CreateCommentRequest true "Comment data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /comments [post]
func (h *InteractionHandler) CreateComment(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var req CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if req.PostID == 0 || req.Content == "" {
		response.BadRequest(c, "postId and content are required")
		return
	}

	var parentID *uint
	if req.ParentID != nil && *req.ParentID > 0 {
		parentID = req.ParentID
	}

	comment, err := h.service.CreateComment(c.Request.Context(), uid, req.PostID, req.Content, parentID)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		if err.Error() == "parent comment not found" {
			response.Error(c, http.StatusNotFound, "parent comment not found")
			return
		}
		if err.Error() == "parent comment does not belong to this post" {
			response.BadRequest(c, "parent comment does not belong to this post")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, comment)
}

// GetPostComments godoc
// @Summary Get comments for a post
// @Tags interactions
// @Param id path int true "Post ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id}/comments [get]
func (h *InteractionHandler) GetPostComments(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.GetPostComments(c.Request.Context(), uint(id), page, limit)
	if err != nil {
		if err.Error() == "post not found" {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// DeleteComment godoc
// @Summary Delete a comment
// @Tags interactions
// @Security Bearer
// @Param id path int true "Comment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /comments/{id} [delete]
func (h *InteractionHandler) DeleteComment(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid comment id")
		return
	}

	err = h.service.DeleteComment(c.Request.Context(), uid, uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "comment not found")
			return
		}
		if err.Error() == "unauthorized: not the comment author" {
			response.Error(c, http.StatusForbidden, "forbidden")
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
