// 本文件是反馈的业务逻辑层：提交反馈(公开给登录用户)、列表与审批(超级管理员后台)。
package service

import (
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

// FeedbackListResult 是反馈分页列表的返回结构。
type FeedbackListResult = PageResult[*model.Feedback]

// FeedbackInput 是提交反馈的入参 DTO。
type FeedbackInput struct {
	Content string
}

// FeedbackService 依赖反馈仓库：提交、分页查询、审批状态流转。
type FeedbackService struct {
	repository *repository.FeedbackRepository
}

// NewFeedbackService 由 router 在装配阶段调用，注入反馈仓库。
func NewFeedbackService(repository *repository.FeedbackRepository) *FeedbackService {
	return &FeedbackService{repository: repository}
}

// isValidFeedbackStatus 判断反馈状态是否为三种合法取值之一。
func isValidFeedbackStatus(status string) bool {
	return status == model.FeedbackStatusPending ||
		status == model.FeedbackStatusApproved ||
		status == model.FeedbackStatusRejected
}

// Submit 提交反馈。业务规则：内容去空白后长度需在 1~1000；新反馈默认待审批。
// 提交人以登录身份(userID)为准，不信任请求体。
func (s *FeedbackService) Submit(userID uint64, input FeedbackInput) (*model.Feedback, error) {
	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 1000 {
		return nil, apperror.ParamError
	}

	feedback := &model.Feedback{
		UserID:  userID,
		Content: input.Content,
		Status:  model.FeedbackStatusPending, // 新提交的反馈默认待审批
	}
	if err := s.repository.Create(feedback); err != nil {
		return nil, err
	}
	return feedback, nil
}

// List 分页查询反馈列表(超管后台)。status 可为空(不过滤)或三种状态之一，按 id 倒序。
func (s *FeedbackService) List(status string, page, pageSize int) (*FeedbackListResult, error) {
	var statuses []string
	if status != "" {
		if !isValidFeedbackStatus(status) {
			return nil, apperror.InvalidFeedbackStatusError
		}
		statuses = []string{status}
	}

	offset := pagination.Offset(page, pageSize)
	feedbacks, total, err := s.repository.GetFeedbacks(statuses, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &FeedbackListResult{
		List:     feedbacks,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Review 审批反馈。业务规则：
//  1. 目标状态只能是 approved 或 rejected(不能改回 pending)；
//  2. 反馈必须存在；
//  3. 只有 pending 的反馈才能被审批(保证“每个反馈只审批一次”)。
func (s *FeedbackService) Review(feedbackID uint64, status string) error {
	if status != model.FeedbackStatusApproved && status != model.FeedbackStatusRejected {
		return apperror.InvalidFeedbackStatusError
	}
	feedback, err := s.repository.GetFeedbackByID(feedbackID)
	if err != nil {
		return err
	}
	if feedback.Status != model.FeedbackStatusPending {
		return apperror.FeedbackNotPendingError
	}
	return s.repository.UpdateFeedbackStatus(feedbackID, status)
}