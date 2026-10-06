package repository

import (
	"errors"

	"LAF/internal/model"
	"LAF/pkg/apperror"

	"gorm.io/gorm"
)

var (
	ErrFeedbackNotFound = errors.New("feedback not found")
)

type FeedbackRepository struct {
	db *gorm.DB
}

func NewFeedbackRepository(db *gorm.DB) *FeedbackRepository {
	return &FeedbackRepository{db: db}
}

// Create 创建反馈
func (r *FeedbackRepository) Create(fb *model.Feedback) error {
	if err := r.db.Create(fb).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// ListAll 管理员获取全部反馈（分页，含状态过滤）
func (r *FeedbackRepository) GetFeedbacks(statuses []string, limit, offset int) ([]*model.Feedback, int64, error) {
	buildQuery := func() *gorm.DB {
		query := r.db.Model(&model.Feedback{})
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		} else {
			query = query.Where("status = ?", "pending")
		}
		return query
	}
	var total int64
	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	var list []*model.Feedback
	err := buildQuery().Order("created_at DESC").Offset(offset).Limit(limit).Find(&list).Error
	if err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return list, total, nil
}

// GetByID 获取单条反馈
func (r *FeedbackRepository) GetByID(id uint64) (*model.Feedback, error) {
	var fb model.Feedback
	err := r.db.Where("id = ?", id).First(&fb).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.FeedbackNotFoundError
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &fb, nil
}

// UpdateStatus 修改反馈状态
func (r *FeedbackRepository) UpdateStatus(id uint64, status string) error {
	if err := r.db.Model(&model.Feedback{}).Where("id = ?", id).
		Update("status", status).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}
