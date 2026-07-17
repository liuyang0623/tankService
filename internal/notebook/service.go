package notebook

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// NotebookService 处理日记本业务逻辑。
type NotebookService struct {
	db *gorm.DB
}

// NewNotebookService 构造。
func NewNotebookService(db *gorm.DB) *NotebookService {
	return &NotebookService{db: db}
}

// EnsureDefault 若用户无任何日记本，自动建一个"默认"本。
func (s *NotebookService) EnsureDefault(ctx context.Context, userID uint) error {
	var count int64
	if err := s.db.Model(&Notebook{}).Where("author_id = ?", userID).Count(&count).Error; err != nil {
		return fmt.Errorf("count notebooks: %w", err)
	}
	if count > 0 {
		return nil
	}
	nb := &Notebook{Name: DefaultNotebookName, Color: "#f0a868", AuthorID: userID}
	if err := s.db.Create(nb).Error; err != nil {
		return fmt.Errorf("create default notebook: %w", err)
	}
	return nil
}

// FindMine 返回当前用户全部日记本（含日记数），created_at ASC（默认本在前）。
func (s *NotebookService) FindMine(ctx context.Context, userID uint) ([]NotebookResponse, error) {
	if err := s.EnsureDefault(ctx, userID); err != nil {
		return nil, err
	}
	var nbs []Notebook
	if err := s.db.Where("author_id = ?", userID).Order("created_at ASC").Find(&nbs).Error; err != nil {
		return nil, fmt.Errorf("find notebooks: %w", err)
	}
	out := make([]NotebookResponse, len(nbs))
	for i, nb := range nbs {
		var cnt int64
		s.db.Table("diaries").Where("notebook_id = ? AND deleted_at IS NULL", nb.ID).Count(&cnt)
		out[i] = NotebookResponse{ID: nb.ID, Name: nb.Name, Color: nb.Color, Cover: nb.Cover, DiaryCount: cnt, CreatedAt: nb.CreatedAt}
	}
	return out, nil
}

// Create 创建日记本。
func (s *NotebookService) Create(ctx context.Context, userID uint, input CreateNotebookInput) (*NotebookResponse, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	color := input.Color
	if color == "" {
		color = "#f0a868"
	}
	nb := &Notebook{Name: input.Name, Color: color, Cover: input.Cover, AuthorID: userID}
	if err := s.db.Create(nb).Error; err != nil {
		return nil, fmt.Errorf("create notebook: %w", err)
	}
	return &NotebookResponse{ID: nb.ID, Name: nb.Name, Color: nb.Color, Cover: nb.Cover, DiaryCount: 0, CreatedAt: nb.CreatedAt}, nil
}

// Update 更新日记本（仅本人）。
func (s *NotebookService) Update(ctx context.Context, id, userID uint, input UpdateNotebookInput) (*NotebookResponse, error) {
	var nb Notebook
	if err := s.db.Where("author_id = ?", userID).First(&nb, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find notebook: %w", err)
	}
	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Color != nil {
		updates["color"] = *input.Color
	}
	if input.Cover != nil {
		updates["cover"] = *input.Cover
	}
	if len(updates) > 0 {
		if err := s.db.Model(&nb).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update notebook: %w", err)
		}
	}
	var cnt int64
	s.db.Table("diaries").Where("notebook_id = ? AND deleted_at IS NULL", nb.ID).Count(&cnt)
	return &NotebookResponse{ID: nb.ID, Name: nb.Name, Color: nb.Color, Cover: nb.Cover, DiaryCount: cnt, CreatedAt: nb.CreatedAt}, nil
}

// Remove 删除日记本（仅本人）。关联日记的 notebook_id 置 0（不级联删日记）。
func (s *NotebookService) Remove(ctx context.Context, id, userID uint) error {
	var nb Notebook
	if err := s.db.Where("author_id = ?", userID).First(&nb, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return gorm.ErrRecordNotFound
		}
		return fmt.Errorf("find notebook: %w", err)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Table("diaries").Where("notebook_id = ?", id).Update("notebook_id", 0).Error; err != nil {
			return err
		}
		return tx.Delete(&Notebook{}, id).Error
	})
}
