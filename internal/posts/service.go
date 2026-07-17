package posts

import (
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// PaginatedResult wraps a paginated list of posts with meta info.
type PaginatedResult struct {
	Data []PostResponse `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// PaginationMeta contains pagination metadata.
type PaginationMeta struct {
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalPages int   `json:"totalPages"`
}

// AuthorInfo is the author representation in the API response.
type AuthorInfo struct {
	ID       uint   `json:"id"`
	Nickname string `json:"name"` // map to "name" for API response
	Avatar   string `json:"avatar"`
}

// PostResponse is the API response format for a single post.
type PostResponse struct {
	ID           uint           `json:"id"`
	Title        string         `json:"title"`
	Content      string         `json:"content"`
	Cover        string         `json:"cover,omitempty"`
	Status       PostStatus     `json:"status"`
	Category     string         `json:"category,omitempty"`
	AuthorID     uint           `json:"authorId"`
	Author       AuthorInfo     `json:"author"`
	ViewCount    int            `json:"viewCount"`
	LikeCount    int            `json:"likeCount"`
	CommentCount int            `json:"commentCount"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	PublishedAt  *time.Time     `json:"publishedAt,omitempty"`
	Images       []ImageInfo    `json:"images,omitempty"`
	Topics       []TopicInfo    `json:"topics,omitempty"`
	IsLiked      bool           `json:"isLiked,omitempty"`
	IsFavorited  bool           `json:"isFavorited,omitempty"`
}

// ImageInfo is a simplified image representation in the API response.
type ImageInfo struct {
	ID        uint   `json:"id"`
	URL       string `json:"url"`
	SortOrder int    `json:"order"`
}

// TopicInfo is a simplified topic representation in the API response.
type TopicInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// PostListResponse is a simplified response for list views.
type PostListResponse struct {
	ID             uint       `json:"id"`
	Title          string     `json:"title"`
	ContentPreview string     `json:"contentPreview"`
	Cover          string     `json:"cover,omitempty"`
	Status         PostStatus `json:"status"`
	Category       string     `json:"category,omitempty"`
	AuthorID       uint       `json:"authorId"`
	Author         AuthorInfo `json:"author"`
	ViewCount      int        `json:"viewCount"`
	LikeCount      int        `json:"likeCount"`
	CommentCount   int        `json:"commentCount"`
	CreatedAt      time.Time  `json:"createdAt"`
	PublishedAt    *time.Time `json:"publishedAt,omitempty"`
	CoverImages    []string   `json:"coverImages,omitempty"`
	TopicNames     []string   `json:"topicNames,omitempty"`
}

// CreatePostInput holds the input for creating a post.
type CreatePostInput struct {
	Title    string     `json:"title"`
	Content  string     `json:"content"`
	Cover    string     `json:"cover"`
	Status   PostStatus `json:"status"`
	Category string     `json:"category"`
	Images   []string   `json:"images"`
	Topics   []string   `json:"topics"`
}

// UpdatePostInput holds the input for updating a post.
type UpdatePostInput struct {
	Title    *string     `json:"title,omitempty"`
	Content  *string     `json:"content,omitempty"`
	Cover    *string     `json:"cover,omitempty"`
	Status   *PostStatus `json:"status,omitempty"`
	Category *string     `json:"category,omitempty"`
	Images   *[]string   `json:"images,omitempty"`
	Topics   *[]string   `json:"topics,omitempty"`
}

// PostService handles post-related business logic.
type PostService struct {
	db *gorm.DB
}

// NewPostService creates a new PostService with the given GORM db.
func NewPostService(db *gorm.DB) *PostService {
	return &PostService{db: db}
}

// Create creates a new post with the given author and fields.
func (s *PostService) Create(ctx context.Context, authorID uint, input CreatePostInput) (*PostResponse, error) {
	if !IsValidCategory(input.Category) {
		return nil, fmt.Errorf("invalid category: %s", input.Category)
	}

	status := input.Status
	if status == "" {
		status = PostStatusDraft
	}

	var publishedAt *time.Time
	if status == PostStatusPublished {
		now := time.Now()
		publishedAt = &now
	}

	post := &Post{
		Title:       input.Title,
		Content:     input.Content,
		Cover:       input.Cover,
		Status:      status,
		Category:    input.Category,
		AuthorID:    authorID,
		PublishedAt: publishedAt,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		// Create images if provided
		if len(input.Images) > 0 {
			images := make([]PostImage, len(input.Images))
			for i, url := range input.Images {
				images[i] = PostImage{
					PostID:    post.ID,
					URL:       url,
					SortOrder: i,
				}
			}
			if err := tx.Create(&images).Error; err != nil {
				return err
			}
			post.Images = images
		}

		// Create topics if provided
		if len(input.Topics) > 0 {
			for _, topicName := range input.Topics {
				name := topicName
				if len(name) > 0 && name[0] == '#' {
					name = name[1:]
				}
				if name == "" {
					continue
				}

				var topic Topic
				if err := tx.Where("name = ?", name).FirstOrCreate(&topic, Topic{Name: name}).Error; err != nil {
					return err
				}

				// Create many2many relation via post_topics table
				if err := tx.Exec("INSERT IGNORE INTO post_topics (post_id, topic_id) VALUES (?, ?)", post.ID, topic.ID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	// Reload with associations
	return s.FindOne(ctx, post.ID, &authorID)
}

// FindAllOptions holds optional filters for the published-post list query.
type FindAllOptions struct {
	Keyword       string // title LIKE filter
	Category      string // exact category filter; "none" = uncategorized (empty/NULL)
	Sort          string // "likes" → order by like_count desc; else published_at desc
	Following     bool   // only posts by authors the current user follows
	CurrentUserID *uint  // required when Following is true
}

// applyFilters builds the shared WHERE clause for FindAll count & find queries.
func applyFilters(q *gorm.DB, opts FindAllOptions) *gorm.DB {
	q = q.Where("status = ?", PostStatusPublished)
	if opts.Keyword != "" {
		q = q.Where("title LIKE ?", "%"+opts.Keyword+"%")
	}
	if opts.Category == "none" {
		// "其他"：无分类的文章（空字符串或 NULL）
		q = q.Where("category = ? OR category IS NULL", "")
	} else if opts.Category != "" {
		q = q.Where("category = ?", opts.Category)
	}
	if opts.Following {
		if opts.CurrentUserID != nil {
			q = q.Where("author_id IN (SELECT following_id FROM follows WHERE follower_id = ?)", *opts.CurrentUserID)
		} else {
			// Following requested without a logged-in user → no results
			q = q.Where("1 = 0")
		}
	}
	return q
}

// FindAll returns a paginated list of published posts with optional filters.
func (s *PostService) FindAll(ctx context.Context, page, limit int, opts FindAllOptions) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := applyFilters(s.db.Model(&Post{}), opts).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count posts: %w", err)
	}

	order := "published_at DESC"
	if opts.Sort == "likes" {
		order = "like_count DESC"
	}

	var posts []Post
	err := applyFilters(s.db, opts).
		Preload("Author").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Limit(3)
		}).
		Preload("Topics").
		Order(order).
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("find all posts: %w", err)
	}

	data := make([]PostResponse, len(posts))
	for i, p := range posts {
		data[i] = s.toPostResponse(p)
	}

	return &PaginatedResult{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// ListCategories returns the fixed category list.
func (s *PostService) ListCategories() []CategoryInfo {
	return FixedCategories
}

// FindOne returns a post by ID, with full associations.
// If currentUserID is provided, also checks isLiked/isFavorited and enforces draft visibility.
func (s *PostService) FindOne(ctx context.Context, id uint, currentUserID *uint) (*PostResponse, error) {
	var post Post
	err := s.db.Preload("Author").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Preload("Topics").
		First(&post, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find post: %w", err)
	}

	// Draft visibility: only the author can view their own drafts
	if post.Status == PostStatusDraft {
		if currentUserID == nil || *currentUserID != post.AuthorID {
			return nil, fmt.Errorf("forbidden: you do not have permission to view this draft")
		}
	}

	// Increment view count
	s.db.Model(&Post{}).Where("id = ?", id).UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	resp := s.toPostResponse(post)
	resp.ViewCount = post.ViewCount + 1 // reflect the increment in response

	// Check isLiked / isFavorited if user is authenticated
	if currentUserID != nil && *currentUserID > 0 {
		var likeCount int64
		s.db.Table("likes").Where("post_id = ? AND user_id = ?", id, *currentUserID).Count(&likeCount)
		resp.IsLiked = likeCount > 0

		var favCount int64
		s.db.Table("favorites").Where("post_id = ? AND user_id = ?", id, *currentUserID).Count(&favCount)
		resp.IsFavorited = favCount > 0
	}

	return &resp, nil
}

// Update updates a post. Only the author can update their own post.
func (s *PostService) Update(ctx context.Context, id, userID uint, input UpdatePostInput) (*PostResponse, error) {
	var post Post
	if err := s.db.First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find post for update: %w", err)
	}

	if post.AuthorID != userID {
		return nil, fmt.Errorf("unauthorized: not the post author")
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{}

		if input.Title != nil {
			updates["title"] = *input.Title
		}
		if input.Content != nil {
			updates["content"] = *input.Content
		}
		if input.Cover != nil {
			updates["cover"] = *input.Cover
		}
		if input.Status != nil {
			updates["status"] = *input.Status
			// If transitioning from DRAFT to PUBLISHED, set published_at
			if *input.Status == PostStatusPublished && post.Status == PostStatusDraft {
				now := time.Now()
				updates["published_at"] = now
			}
		}
		if input.Category != nil {
			if !IsValidCategory(*input.Category) {
				return fmt.Errorf("invalid category: %s", *input.Category)
			}
			updates["category"] = *input.Category
		}

		if len(updates) > 0 {
			if err := tx.Model(&post).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Update images if provided
		if input.Images != nil {
			// Delete existing images
			if err := tx.Where("post_id = ?", id).Delete(&PostImage{}).Error; err != nil {
				return err
			}
			// Create new images
			if len(*input.Images) > 0 {
				images := make([]PostImage, len(*input.Images))
				for i, url := range *input.Images {
					images[i] = PostImage{
						PostID:    id,
						URL:       url,
						SortOrder: i,
					}
				}
				if err := tx.Create(&images).Error; err != nil {
					return err
				}
			}
		}

		// Update topics if provided
		if input.Topics != nil {
			// Delete existing topic relations
			if err := tx.Exec("DELETE FROM post_topics WHERE post_id = ?", id).Error; err != nil {
				return err
			}
			// Create new topic relations
			for _, topicName := range *input.Topics {
				name := topicName
				if len(name) > 0 && name[0] == '#' {
					name = name[1:]
				}
				if name == "" {
					continue
				}

				var topic Topic
				if err := tx.Where("name = ?", name).FirstOrCreate(&topic, Topic{Name: name}).Error; err != nil {
					return err
				}

				if err := tx.Exec("INSERT IGNORE INTO post_topics (post_id, topic_id) VALUES (?, ?)", id, topic.ID).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}

	return s.FindOne(ctx, id, &userID)
}

// Remove deletes a post. Only the author can delete their own post.
func (s *PostService) Remove(ctx context.Context, id, userID uint) error {
	var post Post
	if err := s.db.First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("find post for removal: %w", err)
	}

	if post.AuthorID != userID {
		return fmt.Errorf("unauthorized: not the post author")
	}

	// Delete images and topic relations first
	s.db.Where("post_id = ?", id).Delete(&PostImage{})
	s.db.Exec("DELETE FROM post_topics WHERE post_id = ?", id)

	if err := s.db.Delete(&Post{}, id).Error; err != nil {
		return fmt.Errorf("remove post: %w", err)
	}
	return nil
}

// Publish publishes a draft post.
func (s *PostService) Publish(ctx context.Context, id, userID uint) (*PostResponse, error) {
	var post Post
	if err := s.db.First(&post, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find post for publish: %w", err)
	}

	if post.AuthorID != userID {
		return nil, fmt.Errorf("unauthorized: not the post author")
	}

	if post.Status == PostStatusPublished {
		return nil, fmt.Errorf("post is already published")
	}

	now := time.Now()
	if err := s.db.Model(&post).Updates(map[string]interface{}{
		"status":       PostStatusPublished,
		"published_at": now,
	}).Error; err != nil {
		return nil, fmt.Errorf("publish post: %w", err)
	}

	return s.FindOne(ctx, id, &userID)
}

// FindDrafts returns a paginated list of draft posts for a given author.
func (s *PostService) FindDrafts(ctx context.Context, userID uint, page, limit int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Post{}).Where("author_id = ? AND status = ?", userID, PostStatusDraft).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count drafts: %w", err)
	}

	var posts []Post
	err := s.db.Where("author_id = ? AND status = ?", userID, PostStatusDraft).
		Preload("Author").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Limit(3)
		}).
		Preload("Topics").
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("find drafts: %w", err)
	}

	data := make([]PostResponse, len(posts))
	for i, p := range posts {
		data[i] = s.toPostResponse(p)
	}

	return &PaginatedResult{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// FindUserPosts returns a paginated list of all posts (including drafts) for a given user.
func (s *PostService) FindUserPosts(ctx context.Context, userID uint, page, limit int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.Model(&Post{}).Where("author_id = ?", userID).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count user posts: %w", err)
	}

	var posts []Post
	err := s.db.Where("author_id = ?", userID).
		Preload("Author").
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC").Limit(3)
		}).
		Preload("Topics").
		Order("updated_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&posts).Error
	if err != nil {
		return nil, fmt.Errorf("find user posts: %w", err)
	}

	data := make([]PostResponse, len(posts))
	for i, p := range posts {
		data[i] = s.toPostResponse(p)
	}

	return &PaginatedResult{
		Data: data,
		Meta: PaginationMeta{
			Total:      total,
			Page:       page,
			Limit:      limit,
			TotalPages: int(math.Ceil(float64(total) / float64(limit))),
		},
	}, nil
}

// toPostResponse converts a Post model to a PostResponse DTO.
func (s *PostService) toPostResponse(post Post) PostResponse {
	return ToPostResponse(post)
}

// ToPostResponse converts a Post model to a PostResponse DTO (package-level,
// reusable by other packages such as interactions/favorites).
func ToPostResponse(post Post) PostResponse {
	resp := PostResponse{
		ID:           post.ID,
		Title:        post.Title,
		Content:      post.Content,
		Cover:        post.Cover,
		Status:       post.Status,
		Category:     post.Category,
		AuthorID:     post.AuthorID,
		Author: AuthorInfo{
			ID:       post.Author.ID,
			Nickname: post.Author.Nickname,
			Avatar:   post.Author.Avatar,
		},
		ViewCount:    post.ViewCount,
		LikeCount:    post.LikeCount,
		CommentCount: post.CommentCount,
		CreatedAt:    post.CreatedAt,
		UpdatedAt:    post.UpdatedAt,
		PublishedAt:  post.PublishedAt,
	}

	if len(post.Images) > 0 {
		resp.Images = make([]ImageInfo, len(post.Images))
		for i, img := range post.Images {
			resp.Images[i] = ImageInfo{
				ID:        img.ID,
				URL:       img.URL,
				SortOrder: img.SortOrder,
			}
		}
	}

	if len(post.Topics) > 0 {
		resp.Topics = make([]TopicInfo, len(post.Topics))
		for i, t := range post.Topics {
			resp.Topics[i] = TopicInfo{
				ID:   t.ID,
				Name: t.Name,
			}
		}
	}

	return resp
}
