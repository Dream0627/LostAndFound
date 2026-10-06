package service

import (
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

type FeedbackCreateInput struct {
	UserID  uint64
	Content string `json:"content" binding:"required,min=2,max=500"`
}

type FeedbackHandleInput struct {
	FeedbackID uint64
	status     string
}

type FeedbackItemResult struct {
	ID      uint64 `json:"id"`
	UserID  uint64 `json:"user_id"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type FeedbackService struct {
	repository *repository.FeedbackRepository
}

func NewFeedbackService(repository *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{repository: repository}
}
func IsValidFeedbackStatus(status string) bool {
	return status == model.FeedbackStatusPending || status == model.FeedbackStatusProcessed || status == model.FeedbackStatusRejected
}

// CreateFeedback 用户提交反馈
func (s *FeedbackService) CreateFeedback(input FeedbackCreateInput) (*model.Feedback, error) {
	fb := &model.Feedback{
		UserID:  input.UserID,
		Content: input.Content,
		Status:  model.FeedbackStatusPending,
	}
	err := s.repository.Create(fb)
	if err != nil {
		return fb, apperror.DatabaseError
	}
	return fb, nil
}

type FeedbackListResult = PageResult[*model.Feedback]

// ListFeedback 管理员分页查询反馈
func (s *FeedbackService) GetFeedback(status string, page, pageSize int) (*FeedbackListResult, error) {
	validStatuses := make([]string, 0, len(status))
	if !IsValidFeedbackStatus(status) {
		return nil, apperror.InvalidFeedbackStatusError
	}
	if status != "" {
		validStatuses = append(validStatuses, model.FeedbackStatusPending, model.FeedbackStatusProcessed, model.FeedbackStatusRejected)
	}
	offset := pagination.Offset(page, pageSize)
	feedbacks, total, err := s.repository.GetFeedbacks(validStatuses, pageSize, offset)
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &FeedbackListResult{
		List:     feedbacks,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// HandleFeedback 管理员标记为已处理
func (s *FeedbackService) HandleFeedback(input FeedbackHandleInput) error {
	fb, err := s.repository.GetByID(input.FeedbackID)
	if err != nil {
		if err == repository.ErrFeedbackNotFound {
			return apperror.FeedbackNotFoundError
		}
		return apperror.DatabaseError
	}
	if fb.Status == model.FeedbackStatusProcessed {
		return apperror.FeedbackAlreadyProcessedError
	}
	err = s.repository.UpdateStatus(input.FeedbackID, input.status)
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}
