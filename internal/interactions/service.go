package interactions

import (
	"context"
	"fmt"
	"math"
	"time"

	"go-service/internal/posts"
	"go-service/internal/users"

	"gorm.io/gorm"
)

// CommentAuthor is the author info embedded in a comment response.
type CommentAuthor struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Avatar   string `json:"avatar"`
}

// CommentResponse is the API DTO for a comment (lowercase json fields + author).
type CommentResponse struct {
	ID        uint              `json:"id"`
	Content   string            `json:"content"`
	AuthorID  uint              `json:"authorId"`
	Author    CommentAuthor     `json:"author"`
	ParentID  *uint             `json:"parentId"`
	Replies   []CommentResponse `json:"replies"`
	CreatedAt time.Time         `json:"createdAt"`
}

// PaginatedFavorites wraps a paginated list of favorites with meta info.
type PaginatedFavorites struct {
	Data []FavoriteItem `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// PaginatedComments wraps a paginated list of comment DTOs with meta info.
type PaginatedComments struct {
	Data []CommentResponse `json:"data"`
	Meta PaginationMeta    `json:"meta"`
}

// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// InteractionService handles post interactions (likes, favorites, comments).
type InteractionService struct {
	db *gorm.DB
}

// NewInteractionService creates a new InteractionService.
func NewInteractionService(db *gorm.DB) *InteractionService {
	return &InteractionService{db: db}
}

// LikePost toggles a like on a post. Returns true if liked, false if unliked.
func (s *InteractionService) LikePost(ctx context.Context, userID, postID uint) (bool, error) {
	// Check if post exists
	var postCount int64
	if err := s.db.Model(&posts.Post{}).Where("id = ?", postID).Count(&postCount).Error; err != nil {
		return false, fmt.Errorf("check post: %w", err)
	}
	if postCount == 0 {
		return false, fmt.Errorf("post not found")
	}

	var like Like
	result := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&like)

	if result.Error == nil {
		// 物理删除：Like 有 (post_id,user_id) 唯一索引，软删除会占用索引导致再次点赞唯一键冲突
		if err := s.db.Unscoped().Delete(&like).Error; err != nil {
			return false, fmt.Errorf("remove like: %w", err)
		}
		if err := s.db.Table("posts").Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count - 1")).Error; err != nil {
			return false, fmt.Errorf("decrement like count: %w", err)
		}
		return false, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("check like: %w", result.Error)
	}

	// 清理任何残留（含历史软删除）的同键记录，避免唯一索引冲突
	if err := s.db.Unscoped().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&Like{}).Error; err != nil {
		return false, fmt.Errorf("cleanup like: %w", err)
	}
	like = Like{UserID: userID, PostID: postID}
	if err := s.db.Create(&like).Error; err != nil {
		return false, fmt.Errorf("create like: %w", err)
	}
	if err := s.db.Table("posts").Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
		return false, fmt.Errorf("increment like count: %w", err)
	}
	return true, nil
}

// FavoritePost toggles a favorite on a post. Returns true if favorited, false if unfavorited.
func (s *InteractionService) FavoritePost(ctx context.Context, userID, postID uint) (bool, error) {
	// Check if post exists
	var postCount int64
	if err := s.db.Model(&posts.Post{}).Where("id = ?", postID).Count(&postCount).Error; err != nil {
		return false, fmt.Errorf("check post: %w", err)
	}
	if postCount == 0 {
		return false, fmt.Errorf("post not found")
	}

	var fav Favorite
	result := s.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&fav)

	if result.Error == nil {
		// 物理删除：Favorite 有 (post_id,user_id) 唯一索引，软删除会占用索引导致再次收藏唯一键冲突
		if err := s.db.Unscoped().Delete(&fav).Error; err != nil {
			return false, fmt.Errorf("remove favorite: %w", err)
		}
		return false, nil
	}

	if result.Error != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("check favorite: %w", result.Error)
	}

	// 清理任何残留（含历史软删除）的同键记录，避免唯一索引冲突
	if err := s.db.Unscoped().Where("user_id = ? AND post_id = ?", userID, postID).Delete(&Favorite{}).Error; err != nil {
		return false, fmt.Errorf("cleanup favorite: %w", err)
	}
	fav = Favorite{UserID: userID, PostID: postID}
	if err := s.db.Create(&fav).Error; err != nil {
		return false, fmt.Errorf("create favorite: %w", err)
	}
	return true, nil
}

// FavoriteItem represents a favorited post (as DTO) with its favorite timestamp.
type FavoriteItem struct {
	Post        posts.PostResponse `json:"post"`
	FavoritedAt time.Time          `json:"favoritedAt"`
}

// GetUserFavorites returns the user's favorited posts with pagination.
func (s *InteractionService) GetUserFavorites(ctx context.Context, userID uint, page, limit int) (*PaginatedFavorites, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Favorite{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count favorites: %w", err)
	}

	var favorites []Favorite
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Limit(limit).Offset(offset).Find(&favorites).Error; err != nil {
		return nil, fmt.Errorf("find favorites: %w", err)
	}

	if len(favorites) == 0 {
		return &PaginatedFavorites{
			Data: []FavoriteItem{},
			Meta: PaginationMeta{
				Total:      total,
				Page:       page,
				Limit:      limit,
				TotalPages: int(math.Ceil(float64(total) / float64(limit))),
			},
		}, nil
	}

	postIDs := make([]uint, len(favorites))
	for i, f := range favorites {
		postIDs[i] = f.PostID
	}

	var postList []posts.Post
	if err := s.db.Preload("Author").Preload("Images").Preload("Topics").Where("id IN ?", postIDs).Find(&postList).Error; err != nil {
		return nil, fmt.Errorf("find posts: %w", err)
	}

	postMap := make(map[uint]posts.Post, len(postList))
	for _, p := range postList {
		postMap[p.ID] = p
	}

	result := make([]FavoriteItem, 0, len(favorites))
	for _, f := range favorites {
		if p, ok := postMap[f.PostID]; ok {
			result = append(result, FavoriteItem{
				Post:        posts.ToPostResponse(p),
				FavoritedAt: f.CreatedAt,
			})
		}
	}

	return &PaginatedFavorites{
		Data: result,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// CreateComment creates a new comment on a post.
func (s *InteractionService) CreateComment(ctx context.Context, userID, postID uint, content string, parentID *uint) (*CommentResponse, error) {
	// Check if post exists
	var postCount int64
	if err := s.db.Model(&posts.Post{}).Where("id = ?", postID).Count(&postCount).Error; err != nil {
		return nil, fmt.Errorf("check post: %w", err)
	}
	if postCount == 0 {
		return nil, fmt.Errorf("post not found")
	}

	// If parentID is provided, validate parent comment
	if parentID != nil {
		var parentComment Comment
		if err := s.db.First(&parentComment, *parentID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, fmt.Errorf("parent comment not found")
			}
			return nil, fmt.Errorf("check parent comment: %w", err)
		}
		if parentComment.PostID != postID {
			return nil, fmt.Errorf("parent comment does not belong to this post")
		}
	}

	comment := &Comment{
		PostID:   postID,
		UserID:   userID,
		Content:  content,
		ParentID: parentID,
	}

	if err := s.db.Create(comment).Error; err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	if err := s.db.Table("posts").Where("id = ?", postID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
		return nil, fmt.Errorf("increment comment count: %w", err)
	}

	// 组装作者信息返回 DTO，供前端直接回显
	authors := s.loadAuthors([]uint{userID})
	resp := toCommentResponse(comment, authors)
	return &resp, nil
}

// GetPostComments returns top-level comments for a post with their replies and pagination.
func (s *InteractionService) GetPostComments(ctx context.Context, postID uint, page, limit int) (*PaginatedComments, error) {
	// Check if post exists
	var postCount int64
	if err := s.db.Model(&posts.Post{}).Where("id = ?", postID).Count(&postCount).Error; err != nil {
		return nil, fmt.Errorf("check post: %w", err)
	}
	if postCount == 0 {
		return nil, fmt.Errorf("post not found")
	}

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Comment{}).Where("post_id = ? AND parent_id IS NULL", postID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count comments: %w", err)
	}

	var comments []Comment
	if err := s.db.Where("post_id = ? AND parent_id IS NULL", postID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&comments).Error; err != nil {
		return nil, fmt.Errorf("find comments: %w", err)
	}

	for i := range comments {
		if err := s.db.Where("parent_id = ?", comments[i].ID).Order("created_at ASC").Find(&comments[i].Replies).Error; err != nil {
			return nil, fmt.Errorf("find replies: %w", err)
		}
	}

	// 组装作者信息并转为 DTO
	authors := s.loadAuthors(collectUserIDs(comments))
	data := make([]CommentResponse, 0, len(comments))
	for i := range comments {
		data = append(data, toCommentResponse(&comments[i], authors))
	}

	return &PaginatedComments{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// collectUserIDs gathers all distinct user IDs from comments and their replies.
func collectUserIDs(comments []Comment) []uint {
	seen := map[uint]struct{}{}
	var walk func(cs []Comment)
	walk = func(cs []Comment) {
		for i := range cs {
			seen[cs[i].UserID] = struct{}{}
			if len(cs[i].Replies) > 0 {
				walk(cs[i].Replies)
			}
		}
	}
	walk(comments)
	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	return ids
}

// loadAuthors fetches users by IDs into a lookup map.
func (s *InteractionService) loadAuthors(ids []uint) map[uint]CommentAuthor {
	result := map[uint]CommentAuthor{}
	if len(ids) == 0 {
		return result
	}
	var us []users.User
	if err := s.db.Where("id IN ?", ids).Find(&us).Error; err != nil {
		return result
	}
	for i := range us {
		result[us[i].ID] = CommentAuthor{ID: us[i].ID, Name: us[i].Nickname, Avatar: us[i].Avatar}
	}
	return result
}

// toCommentResponse converts a Comment model (with replies) to a DTO recursively.
func toCommentResponse(c *Comment, authors map[uint]CommentAuthor) CommentResponse {
	replies := make([]CommentResponse, 0, len(c.Replies))
	for i := range c.Replies {
		replies = append(replies, toCommentResponse(&c.Replies[i], authors))
	}
	return CommentResponse{
		ID:        c.ID,
		Content:   c.Content,
		AuthorID:  c.UserID,
		Author:    authors[c.UserID],
		ParentID:  c.ParentID,
		Replies:   replies,
		CreatedAt: c.CreatedAt,
	}
}

// DeleteComment deletes a comment and its replies, decrementing the post's comment count.
func (s *InteractionService) DeleteComment(ctx context.Context, userID, commentID uint) error {
	var comment Comment
	if err := s.db.First(&comment, commentID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("find comment: %w", err)
	}

	if comment.UserID != userID {
		return fmt.Errorf("unauthorized: not the comment author")
	}

	var replyCount int64
	if err := s.db.Model(&Comment{}).Where("parent_id = ?", commentID).Count(&replyCount).Error; err != nil {
		return fmt.Errorf("count replies: %w", err)
	}

	if err := s.db.Delete(&comment).Error; err != nil {
		return fmt.Errorf("delete comment: %w", err)
	}

	totalDeleted := int(replyCount) + 1
	if err := s.db.Table("posts").Where("id = ?", comment.PostID).UpdateColumn("comment_count", gorm.Expr("comment_count - ?", totalDeleted)).Error; err != nil {
		return fmt.Errorf("decrement comment count: %w", err)
	}

	return nil
}
