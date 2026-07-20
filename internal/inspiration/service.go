package inspiration

import (
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

// ===================== 解惑问答 Service =====================

// QAService 处理解惑问答的业务逻辑。
type QAService struct {
	db *gorm.DB
}

// NewQAService 创建 QAService。
func NewQAService(db *gorm.DB) *QAService {
	return &QAService{db: db}
}

// CreateQuestion 创建一个问题，关联提问者。
func (s *QAService) CreateQuestion(ctx context.Context, userID uint, input CreateQuestionInput) (*QuestionResponse, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	q := &Question{
		AuthorID: userID,
		Title:    input.Title,
		Content:  input.Content,
	}
	if err := s.db.WithContext(ctx).Create(q).Error; err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}

	return &QuestionResponse{
		ID:          q.ID,
		AuthorID:    q.AuthorID,
		Title:       q.Title,
		Content:     q.Content,
		AnswerCount: 0,
		Answers:     []AnswerResponse{},
		CreatedAt:   q.CreatedAt,
		UpdatedAt:   q.UpdatedAt,
	}, nil
}

// ListQuestions 返回全站问题的分页列表（倒序），每项含回答数。
func (s *QAService) ListQuestions(ctx context.Context, page, limit int) (*PaginatedResult, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	var total int64
	if err := s.db.WithContext(ctx).Model(&Question{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count questions: %w", err)
	}

	var questions []Question
	if err := s.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&questions).Error; err != nil {
		return nil, fmt.Errorf("find questions: %w", err)
	}

	// 一次聚合查询获取每个问题的回答数，避免 N+1。
	counts := s.answerCounts(ctx, questionIDs(questions))

	data := make([]QuestionListItem, len(questions))
	for i, q := range questions {
		data[i] = QuestionListItem{
			ID:          q.ID,
			AuthorID:    q.AuthorID,
			Title:       q.Title,
			AnswerCount: counts[q.ID],
			CreatedAt:   q.CreatedAt,
		}
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

// GetQuestion 返回问题详情与全部回答（正序）。
func (s *QAService) GetQuestion(ctx context.Context, id uint) (*QuestionResponse, error) {
	var q Question
	err := s.db.WithContext(ctx).
		Preload("Answers", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		First(&q, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find question: %w", err)
	}

	answers := make([]AnswerResponse, len(q.Answers))
	for i, a := range q.Answers {
		answers[i] = AnswerResponse{
			ID:        a.ID,
			AuthorID:  a.AuthorID,
			Content:   a.Content,
			CreatedAt: a.CreatedAt,
		}
	}

	return &QuestionResponse{
		ID:          q.ID,
		AuthorID:    q.AuthorID,
		Title:       q.Title,
		Content:     q.Content,
		AnswerCount: len(answers),
		Answers:     answers,
		CreatedAt:   q.CreatedAt,
		UpdatedAt:   q.UpdatedAt,
	}, nil
}

// CreateAnswer 为某问题创建一条回答。任意用户可回答任意问题。
func (s *QAService) CreateAnswer(ctx context.Context, questionID, userID uint, input CreateAnswerInput) (*AnswerResponse, error) {
	if input.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// 校验问题存在。
	var q Question
	if err := s.db.WithContext(ctx).First(&q, questionID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find question for answer: %w", err)
	}

	a := &Answer{
		QuestionID: questionID,
		AuthorID:   userID,
		Content:    input.Content,
	}
	if err := s.db.WithContext(ctx).Create(a).Error; err != nil {
		return nil, fmt.Errorf("create answer: %w", err)
	}

	return &AnswerResponse{
		ID:        a.ID,
		AuthorID:  a.AuthorID,
		Content:   a.Content,
		CreatedAt: a.CreatedAt,
	}, nil
}

// answerCounts 返回 questionID -> 回答数 的映射。
func (s *QAService) answerCounts(ctx context.Context, ids []uint) map[uint]int {
	result := make(map[uint]int, len(ids))
	if len(ids) == 0 {
		return result
	}

	type row struct {
		QuestionID uint
		Cnt        int
	}
	var rows []row
	s.db.WithContext(ctx).
		Model(&Answer{}).
		Select("question_id, count(*) as cnt").
		Where("question_id IN ?", ids).
		Group("question_id").
		Scan(&rows)

	for _, r := range rows {
		result[r.QuestionID] = r.Cnt
	}
	return result
}

// questionIDs 提取问题 ID 列表。
func questionIDs(questions []Question) []uint {
	ids := make([]uint, len(questions))
	for i, q := range questions {
		ids[i] = q.ID
	}
	return ids
}

// ===================== 运动计划 Service =====================

// SportService 处理运动计划的业务逻辑。
type SportService struct {
	db *gorm.DB
}

// NewSportService 创建 SportService。
func NewSportService(db *gorm.DB) *SportService {
	return &SportService{db: db}
}

// CreateGoal 创建运动目标，初始进度为 0。
func (s *SportService) CreateGoal(ctx context.Context, userID uint, input CreateSportGoalInput) (*SportGoalResponse, error) {
	if input.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	goal := &SportGoal{
		UserID:     userID,
		Name:       input.Name,
		Type:       input.Type,
		Icon:       input.Icon,
		TargetDays: input.TargetDays,
	}
	if err := s.db.WithContext(ctx).Create(goal).Error; err != nil {
		return nil, fmt.Errorf("create sport goal: %w", err)
	}

	return s.toGoalResponse(*goal, today()), nil
}

// ListGoals 返回当前用户的运动目标列表（倒序）。
func (s *SportService) ListGoals(ctx context.Context, userID uint) ([]SportGoalResponse, error) {
	var goals []SportGoal
	if err := s.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&goals).Error; err != nil {
		return nil, fmt.Errorf("list sport goals: %w", err)
	}

	t := today()
	data := make([]SportGoalResponse, len(goals))
	for i, g := range goals {
		data[i] = *s.toGoalResponse(g, t)
	}
	return data, nil
}

// UpdateGoal 更新运动目标，仅所有者可更新。
func (s *SportService) UpdateGoal(ctx context.Context, id, userID uint, input UpdateSportGoalInput) (*SportGoalResponse, error) {
	var goal SportGoal
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&goal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find goal for update: %w", err)
	}

	updates := map[string]interface{}{}
	if input.Name != nil {
		updates["name"] = *input.Name
	}
	if input.Type != nil {
		updates["type"] = *input.Type
	}
	if input.Icon != nil {
		updates["icon"] = *input.Icon
	}
	if input.TargetDays != nil {
		updates["target_days"] = *input.TargetDays
	}
	if len(updates) > 0 {
		if err := s.db.WithContext(ctx).Model(&goal).Updates(updates).Error; err != nil {
			return nil, fmt.Errorf("update goal: %w", err)
		}
	}

	return s.toGoalResponse(goal, today()), nil
}

// Checkin 为目标进行当日打卡。同日重复打卡幂等，连续天数按规则更新。
func (s *SportService) Checkin(ctx context.Context, id, userID uint) (*CheckinResponse, error) {
	var goal SportGoal
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&goal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, fmt.Errorf("find goal for checkin: %w", err)
	}

	t := today()
	awarded := false

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 同日是否已有记录 → 幂等。
		var existing int64
		if err := tx.Model(&SportRecord{}).
			Where("goal_id = ? AND checkin_date = ?", goal.ID, t).
			Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil // 幂等：进度不变。
		}

		// 新建当日打卡记录。
		rec := &SportRecord{GoalID: goal.ID, UserID: userID, CheckinDate: t}
		if err := tx.Create(rec).Error; err != nil {
			return err
		}

		// 计算新的连续天数并更新冗余字段。
		goal.Streak = computeStreakFrom(goal.LastCheckinDate, t, goal.Streak)
		goal.TotalDays += 1
		last := t
		goal.LastCheckinDate = &last
		if err := tx.Model(&SportGoal{}).
			Where("id = ?", goal.ID).
			Updates(map[string]interface{}{
				"streak":            goal.Streak,
				"total_days":        goal.TotalDays,
				"last_checkin_date": goal.LastCheckinDate,
			}).Error; err != nil {
			return err
		}
		awarded = true
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("checkin: %w", err)
	}

	return &CheckinResponse{
		GoalID:         goal.ID,
		Streak:         goal.Streak,
		TotalDays:      goal.TotalDays,
		CheckedInToday: true,
		Awarded:        awarded,
	}, nil
}

// toGoalResponse 将 SportGoal 转为响应 DTO，并根据当日判断是否已打卡。
func (s *SportService) toGoalResponse(g SportGoal, t time.Time) *SportGoalResponse {
	checkedToday := g.LastCheckinDate != nil && sameDay(*g.LastCheckinDate, t)
	return &SportGoalResponse{
		ID:             g.ID,
		Name:           g.Name,
		Type:           g.Type,
		Icon:           g.Icon,
		TargetDays:     g.TargetDays,
		Streak:         g.Streak,
		TotalDays:      g.TotalDays,
		CheckedInToday: checkedToday,
		LastCheckinAt:  g.LastCheckinDate,
		CreatedAt:      g.CreatedAt,
	}
}

// ===================== 纯函数（可单测） =====================

// computeStreakFrom 依据上次打卡日、今天、当前连续天数，计算打卡后的连续天数。
//   - last 为 nil（首次）→ 1
//   - last == 今天 → 不应发生（幂等已拦截），保持 currentStreak（至少为 1）
//   - last == 昨天（连续）→ currentStreak + 1
//   - 其他（漏打，间隔 > 1 天）→ 重置为 1
//
// 不读取数据库、不依赖包级状态，纯函数便于并发安全与单元测试。
func computeStreakFrom(last *time.Time, todayDate time.Time, currentStreak int) int {
	if last == nil {
		return 1
	}
	l := truncateDay(*last)
	td := truncateDay(todayDate)
	diff := int(td.Sub(l).Hours() / 24)
	switch diff {
	case 0:
		if currentStreak < 1 {
			return 1
		}
		return currentStreak
	case 1:
		return currentStreak + 1
	default:
		return 1
	}
}

// today 返回服务器本地当前自然日（00:00）。
func today() time.Time {
	return truncateDay(time.Now())
}

// truncateDay 截断到自然日 00:00（本地时区）。
func truncateDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// sameDay 判断两个时间是否为同一自然日。
func sameDay(a, b time.Time) bool {
	return truncateDay(a).Equal(truncateDay(b))
}
