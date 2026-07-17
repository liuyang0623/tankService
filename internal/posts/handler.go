package posts

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"go-service/pkg/response"
)

// postServiceIface abstracts PostService for handler injection and testing.
type postServiceIface interface {
	Create(ctx context.Context, authorID uint, input CreatePostInput) (*PostResponse, error)
	FindAll(ctx context.Context, page, limit int, opts FindAllOptions) (*PaginatedResult, error)
	FindOne(ctx context.Context, id uint, currentUserID *uint) (*PostResponse, error)
	Update(ctx context.Context, id, userID uint, input UpdatePostInput) (*PostResponse, error)
	Remove(ctx context.Context, id, userID uint) error
	Publish(ctx context.Context, id, userID uint) (*PostResponse, error)
	FindDrafts(ctx context.Context, userID uint, page, limit int) (*PaginatedResult, error)
	FindUserPosts(ctx context.Context, userID uint, page, limit int) (*PaginatedResult, error)
	ListCategories() []CategoryInfo
}

// PostHandler handles HTTP requests for posts.
type PostHandler struct {
	service postServiceIface
}

// NewPostHandler creates a new PostHandler with the given service.
func NewPostHandler(service *PostService) *PostHandler {
	return &PostHandler{service: service}
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

// optionalUserID retrieves the userID if present (non-authenticated endpoints).
func optionalUserID(c *gin.Context) *uint {
	val, ok := c.Get("userID")
	if !ok {
		return nil
	}
	uid, ok := val.(uint)
	if !ok {
		return nil
	}
	return &uid
}

// Create godoc
// @Summary Create a new post
// @Tags posts
// @Security Bearer
// @Accept json
// @Produce json
// @Param body body CreatePostInput true "Post data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts [post]
func (h *PostHandler) Create(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var input CreatePostInput
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

	post, err := h.service.Create(c.Request.Context(), uid, input)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, post)
}

// FindAll godoc
// @Summary Get a paginated list of published posts
// @Tags posts
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Param keyword query string false "Title keyword search"
// @Param category query string false "Category filter"
// @Param sort query string false "Sort mode (likes)"
// @Param following query bool false "Only followed authors' posts"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts [get]
func (h *PostHandler) FindAll(c *gin.Context) {
	page, limit := parsePagination(c)

	opts := FindAllOptions{
		Keyword:  c.Query("keyword"),
		Category: c.Query("category"),
		Sort:     c.Query("sort"),
		Following: c.Query("following") == "true",
	}
	// following 需要当前用户 id（OptionalJWT，未登录时忽略）
	if opts.Following {
		if uid, ok := getUserID(c); ok {
			opts.CurrentUserID = &uid
		}
	}

	result, err := h.service.FindAll(c.Request.Context(), page, limit, opts)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// ListCategories godoc
// @Summary Get the fixed post category list
// @Tags posts
// @Success 200 {object} map[string]interface{}
// @Router /categories [get]
func (h *PostHandler) ListCategories(c *gin.Context) {
	response.Success(c, h.service.ListCategories())
}

// FindOne godoc
// @Summary Get a post by ID
// @Tags posts
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id} [get]
func (h *PostHandler) FindOne(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid post id")
		return
	}

	currentUserID := optionalUserID(c)

	post, err := h.service.FindOne(c.Request.Context(), uint(id), currentUserID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		if err.Error() == "forbidden: you do not have permission to view this draft" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, post)
}

// Update godoc
// @Summary Update a post
// @Tags posts
// @Security Bearer
// @Accept json
// @Produce json
// @Param id path int true "Post ID"
// @Param body body UpdatePostInput true "Update fields"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id} [patch]
func (h *PostHandler) Update(c *gin.Context) {
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

	var input UpdatePostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	post, err := h.service.Update(c.Request.Context(), uint(id), uid, input)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		if err.Error() == "unauthorized: not the post author" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, post)
}

// Remove godoc
// @Summary Delete a post
// @Tags posts
// @Security Bearer
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id} [delete]
func (h *PostHandler) Remove(c *gin.Context) {
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

	err = h.service.Remove(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		if err.Error() == "unauthorized: not the post author" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, gin.H{"message": "deleted"})
}

// Publish godoc
// @Summary Publish a draft post
// @Tags posts
// @Security Bearer
// @Param id path int true "Post ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/{id}/publish [post]
func (h *PostHandler) Publish(c *gin.Context) {
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

	post, err := h.service.Publish(c.Request.Context(), uint(id), uid)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, "post not found")
			return
		}
		if err.Error() == "unauthorized: not the post author" {
			response.Error(c, http.StatusForbidden, "forbidden")
			return
		}
		if err.Error() == "post is already published" {
			response.Error(c, http.StatusForbidden, "post is already published")
			return
		}
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, post)
}

// FindDrafts godoc
// @Summary Get current user's draft posts
// @Tags posts
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/drafts [get]
func (h *PostHandler) FindDrafts(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.FindDrafts(c.Request.Context(), uid, page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// FindUserPosts godoc
// @Summary Get all posts by user ID (including drafts for own user)
// @Tags posts
// @Param id path int true "User ID"
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /users/{id}/posts [get]
func (h *PostHandler) FindUserPosts(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.FindUserPosts(c.Request.Context(), uint(userID), page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
}

// FindMyPosts godoc
// @Summary Get current user's all posts (including drafts)
// @Tags posts
// @Security Bearer
// @Param page query int false "Page number"
// @Param limit query int false "Page size"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /posts/my [get]
func (h *PostHandler) FindMyPosts(c *gin.Context) {
	uid, ok := getUserID(c)
	if !ok {
		response.Unauthorized(c, "unauthorized")
		return
	}

	page, limit := parsePagination(c)

	result, err := h.service.FindUserPosts(c.Request.Context(), uid, page, limit)
	if err != nil {
		response.InternalError(c, "server error")
		return
	}

	response.Success(c, result)
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
