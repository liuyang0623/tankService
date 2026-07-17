package diary

import (
	"context"
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
)

// DiaryService handles diary-related business logic.
type DiaryService struct {
	db *gorm.DB
}

// NewDiaryService creates a new DiaryService with the given GORM db.
func NewDiaryService(db *gorm.DB) *DiaryService {
	return &DiaryService{db: db}
}

// Create creates a new diary entry for the given user.
func (s *DiaryService) Create(ctx context.Context, userID uint, input CreateDiaryInput) (*DiaryResponse, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}
	if input.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	diary := &Diary{
		Title:    input.Title,
		Content:  input.Content,
		Cover:    input.Cover,
		Mood:     input.Mood,
		Weather:  input.Weather,
		NotebookID: input.NotebookID,
		AuthorID: userID,
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(diary).Error; err != nil {
			return err
		}

		// Create images if provided
		if len(input.Images) > 0 {
			images := make([]DiaryImage, len(input.Images))
			for i, url := range input.Images {
				images[i] = DiaryImage{
					DiaryID:   diary.ID,
					URL:       url,
					SortOrder: i,
				}
			}
			if err := tx.Create(&images).Error; err != nil {
				return err
			}
			diary.Images = images
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("create diary: %w", err)
	}

	return s.toDiaryResponse(dbDiary(diary)), nil
}

// FindMine returns a paginated list of the current user's diary entries.
// notebookID > 0 filters to a specific notebook; 0 returns all.
func (s *DiaryService) FindMine(ctx context.Context, userID uint, notebookID uint, page, limit int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	base := s.db.Model(&Diary{}).Where("author_id = ?", userID)
	if notebookID > 0 {
		base = base.Where("notebook_id = ?", notebookID)
	}

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count diaries: %w", err)
	}

	q := s.db.Where("author_id = ?", userID)
	if notebookID > 0 {
		q = q.Where("notebook_id = ?", notebookID)
	}
	var diaries []Diary
	err := q.
		Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&diaries).Error
	if err != nil {
		return nil, fmt.Errorf("find diaries: %w", err)
	}

	data := make([]DiaryListItem, len(diaries))
	for i, d := range diaries {
		data[i] = s.toListItem(d)
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

// FindOne returns a single diary entry by ID, only if it belongs to the user.
func (s *DiaryService) FindOne(ctx context.Context, id, userID uint) (*DiaryResponse, error) {
	var diary Diary
	err := s.db.Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("author_id = ?", userID).First(&diary, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find diary: %w", err)
	}

	return s.toDiaryResponse(diary), nil
}

// Update updates a diary entry. Only the author can update their own diary.
func (s *DiaryService) Update(ctx context.Context, id, userID uint, input UpdateDiaryInput) (*DiaryResponse, error) {
	var diary Diary
	if err := s.db.Where("author_id = ?", userID).First(&diary, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find diary for update: %w", err)
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
		if input.Mood != nil {
			updates["mood"] = *input.Mood
		}
		if input.Weather != nil {
			updates["weather"] = *input.Weather
		}
		if input.NotebookID != nil {
			updates["notebook_id"] = *input.NotebookID
		}

		if len(updates) > 0 {
			if err := tx.Model(&diary).Updates(updates).Error; err != nil {
				return err
			}
		}

		// Update images if provided
		if input.Images != nil {
			// Delete existing images
			if err := tx.Where("diary_id = ?", id).Delete(&DiaryImage{}).Error; err != nil {
				return err
			}
			// Create new images
			if len(*input.Images) > 0 {
				images := make([]DiaryImage, len(*input.Images))
				for i, url := range *input.Images {
					images[i] = DiaryImage{
						DiaryID:   id,
						URL:       url,
						SortOrder: i,
					}
				}
				if err := tx.Create(&images).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("update diary: %w", err)
	}

	return s.FindOne(ctx, id, userID)
}

// Remove deletes a diary entry. Only the author can delete their own diary.
func (s *DiaryService) Remove(ctx context.Context, id, userID uint) error {
	var diary Diary
	if err := s.db.Where("author_id = ?", userID).First(&diary, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("find diary for removal: %w", err)
	}

	// Delete images first (though cascade should handle it)
	s.db.Where("diary_id = ?", id).Delete(&DiaryImage{})

	if err := s.db.Delete(&Diary{}, id).Error; err != nil {
		return fmt.Errorf("remove diary: %w", err)
	}
	return nil
}

// toDiaryResponse converts a Diary model to a DiaryResponse DTO.
func (s *DiaryService) toDiaryResponse(diary Diary) *DiaryResponse {
	resp := &DiaryResponse{
		ID:        diary.ID,
		Title:     diary.Title,
		Content:   diary.Content,
		Cover:     diary.Cover,
		Mood:      diary.Mood,
		Weather:   diary.Weather,
		NotebookID: diary.NotebookID,
		CreatedAt: diary.CreatedAt,
		UpdatedAt: diary.UpdatedAt,
	}

	if len(diary.Images) > 0 {
		resp.Images = make([]ImageInfo, len(diary.Images))
		for i, img := range diary.Images {
			resp.Images[i] = ImageInfo{
				ID:        img.ID,
				URL:       img.URL,
				SortOrder: img.SortOrder,
			}
		}
	}

	return resp
}

// toListItem converts a Diary model to a DiaryListItem DTO.
func (s *DiaryService) toListItem(diary Diary) DiaryListItem {
	item := DiaryListItem{
		ID:        diary.ID,
		Title:     diary.Title,
		Cover:     diary.Cover,
		Mood:      diary.Mood,
		Weather:   diary.Weather,
		NotebookID: diary.NotebookID,
		CreatedAt: diary.CreatedAt,
	}

	// Build a text-only preview from the first N chars of content
	preview := stripHTMLTags(diary.Content)
	if len([]rune(preview)) > 100 {
		preview = string([]rune(preview)[:100]) + "..."
	}
	item.ContentPreview = preview

	return item
}

// stripHTMLTags strips HTML tags for content preview.
func stripHTMLTags(s string) string {
	var buf strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}
	return buf.String()
}

// dbDiary returns a shallow copy of the Diary model fields for DTO conversion.
// This helper avoids needing to reload the diary after creation.
func dbDiary(d *Diary) Diary {
	return Diary{
		Model:     gorm.Model{ID: d.ID, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt},
		Title:     d.Title,
		Content:   d.Content,
		Cover:     d.Cover,
		Mood:      d.Mood,
		Weather:   d.Weather,
		NotebookID: d.NotebookID,
		AuthorID:  d.AuthorID,
		Images:    d.Images,
	}
}