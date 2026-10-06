// 本文件是反馈的数据访问层：创建反馈、按主键查询、更新审批状态、按状态分页查询。
// 与其它 repository 一致：只和数据库打交道，并把数据库错误翻译成领域错误(apperror)向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// FeedbackRepository 持有数据库句柄 db，为反馈提供数据访问能力。
type FeedbackRepository struct {
	db *gorm.DB
}

// NewFeedbackRepository 是构造函数，由 router 在装配时注入 db。
func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create 插入一条反馈，自增主键会回填到 feedback.ID。
func (r *FeedbackRepository) Create(feedback *model.Feedback) error {
	if err := r.db.Create(feedback).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetFeedbackByID 按主键查询一条(未删除的)反馈；查不到时返回 apperror.FeedbackNotFoundError。
func (r *FeedbackRepository) GetFeedbackByID(feedbackID uint64) (*model.Feedback, error) {
	var feedback model.Feedback
	err := r.db.Where("id = ?", feedbackID).First(&feedback).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.FeedbackNotFoundError
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &feedback, nil
}

// UpdateFeedbackStatus 更新反馈的审批状态(如 pending -> approved / rejected)。
func (r *FeedbackRepository) UpdateFeedbackStatus(feedbackID uint64, status string) error {
	if err := r.db.Model(&model.Feedback{}).Where("id = ?", feedbackID).Update("status", status).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetFeedbacks 分页查询反馈，可按状态过滤(statuses 为空则不过滤)，按 id 倒序(新反馈在前)，并回填提交人姓名。
// 用闭包复用“同一套过滤条件”：先 Count 求总数，再 Limit/Offset 取当页数据。
func (r *FeedbackRepository) GetFeedbacks(statuses []string, limit, offset int) ([]*model.Feedback, int64, error) {
	var feedbacks []*model.Feedback
	var total int64

	buildQuery := func() *gorm.DB {
		query := r.db.Model(&model.Feedback{})
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
		return query
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&feedbacks).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	r.fillFeedbackAuthorNames(feedbacks)
	return feedbacks, total, nil
}

// fillFeedbackAuthorNames 为一批反馈批量填充提交人姓名(AuthorName，非表字段)。
func (r *FeedbackRepository) fillFeedbackAuthorNames(feedbacks []*model.Feedback) {
	if len(feedbacks) == 0 {
		return
	}
	ids := make([]uint64, 0, len(feedbacks))
	for _, f := range feedbacks {
		if f != nil {
			ids = append(ids, f.UserID)
		}
	}
	names := loadUserNames(r.db, collectUserIDs(ids))
	for _, f := range feedbacks {
		if f == nil {
			continue
		}
		if name, ok := names[f.UserID]; ok && name != "" {
			f.AuthorName = name
		} else {
			f.AuthorName = authorNameUnknown
		}
	}
}