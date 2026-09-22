// 本文件是超级管理员(mainadmin)的业务逻辑：注销/恢复用户、审核申诉、汇总待审批请求。
package service

import (
	"errors"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

// 待审批列表支持的“请求类型”。post 表示帖子发布待审核，appeal 表示注销申诉待审核。
const (
	ReviewTypePost   = "post"
	ReviewTypeAppeal = "appeal"
)

// MainAdminService 依赖三个仓库：
//   - userRepository   ：注销/恢复用户(含同批内容)；
//   - postRepository   ：汇总“待审核帖子”；
//   - appealRepository ：审核申诉、汇总“待审核申诉”。
type MainAdminService struct {
	userRepository   *repository.UserRepository
	postRepository   *repository.PostRepository
	appealRepository *repository.AppealRepository
}

// NewMainAdminService 由 router 注入用户、帖子、申诉三个仓库。
func NewMainAdminService(userRepository *repository.UserRepository, postRepository *repository.PostRepository, appealRepository *repository.AppealRepository) *MainAdminService {
	return &MainAdminService{
		userRepository:   userRepository,
		postRepository:   postRepository,
		appealRepository: appealRepository,
	}
}

// DeleteUser 注销指定用户(管理员入口)。规则：
//  1. 用户必须存在(含已注销，用 Unscoped 查询)；
//  2. 若用户已处于注销状态，返回“已注销”错误，避免重复注销破坏“同批时间戳”的语义；
//  3. 注销会连带软删除该用户名下的帖子与评论(仓库层事务内完成)。
func (s *MainAdminService) DeleteUser(userID uint64) error {
	user, err := s.userRepository.GetUserByIDUnscoped(userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return apperror.UserNotFoundError
	}
	if err != nil {
		return err
	}
	if user.DeletedAt.Valid {
		return apperror.UserAlreadyDeactivatedError
	}
	return s.userRepository.DeactivateUser(userID)
}

// RecoverUser 恢复被注销的用户，并“仅恢复与其同批删除的帖子与评论”。
// 若用户不存在返回 404；若用户本就未被注销返回“未注销”错误(无可恢复)。
func (s *MainAdminService) RecoverUser(userID uint64) error {
	err := s.userRepository.RecoverUser(userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return apperror.UserNotFoundError
	}
	if errors.Is(err, repository.ErrUserNotDeleted) {
		return apperror.UserNotDeactivatedError
	}
	return err
}

// ReviewAppeal 审核申诉。规则：
//  1. 目标状态只能是 approved 或 rejected(不能改回 pending)；
//  2. 申诉必须存在；
//  3. 只有 pending 的申诉才能被审核(保证“每个申诉只审核一次”)；
//  4. 审核为 approved 时，自动级联恢复该申诉对应的账号(及其同批内容)。
//
// 说明：状态更新与账号恢复分属两个仓库、非同一事务；若“申诉已通过”已落库而恢复阶段报错会向上抛出，
// 属于可接受的最终一致(教学项目从简；生产可用同一数据库事务或补偿机制保证强一致)。
func (s *MainAdminService) ReviewAppeal(appealID uint64, status string) error {
	if status != model.AppealStatusApproved && status != model.AppealStatusRejected {
		return apperror.InvalidAppealStatusError
	}

	appeal, err := s.appealRepository.GetAppealByID(appealID)
	if errors.Is(err, repository.ErrAppealNotFound) {
		return apperror.AppealNotFoundError
	}
	if err != nil {
		return err
	}
	if appeal.Status != model.AppealStatusPending {
		return apperror.AppealNotPendingError
	}

	if err := s.appealRepository.UpdateAppealStatus(appealID, status); err != nil {
		return err
	}

	if status == model.AppealStatusApproved {
		// 自动恢复账号；若账号本就未被注销(例如已被管理员手动恢复)，视为已达成目标，不报错。
		if err := s.RecoverUser(appeal.UserID); err != nil && !errors.Is(err, apperror.UserNotDeactivatedError) {
			return err
		}
	}
	return nil
}

// PendingReviewResult 是“待审批列表”的返回结构。
// 帖子与申诉结构差异较大，故分开成两个数组返回；type 过滤只会填充其中一个。
type PendingReviewResult struct {
	Posts   []*model.Post   `json:"posts"`
	Appeals []*model.Appeal `json:"appeals"`
}

// GetPendingReviews 汇总“待审批请求”。默认(type 为空)同时返回“待审核帖子”和“待审核申诉”；
// 传 type=post 只返回待审核帖子，传 type=appeal 只返回待审核申诉；其它取值返回参数错误。
// 两个列表各自按 page/page_size 分页(offset = (page-1)*pageSize)，均按 id 倒序。
func (s *MainAdminService) GetPendingReviews(reviewType string, page, pageSize int) (*PendingReviewResult, error) {
	includePosts := reviewType == "" || reviewType == ReviewTypePost
	includeAppeals := reviewType == "" || reviewType == ReviewTypeAppeal
	if !includePosts && !includeAppeals {
		return nil, apperror.InvalidReviewTypeError
	}

	offset := (page - 1) * pageSize
	result := &PendingReviewResult{
		Posts:   make([]*model.Post, 0),
		Appeals: make([]*model.Appeal, 0),
	}

	if includePosts {
		posts, _, err := s.postRepository.GetPosts(nil, []string{model.PostStatusPending}, nil, pageSize, offset) // finished=nil: 待审核列表不限完成状态
		if err != nil {
			return nil, err
		}
		result.Posts = posts
	}

	if includeAppeals {
		appeals, _, err := s.appealRepository.GetAppeals([]string{model.AppealStatusPending}, pageSize, offset)
		if err != nil {
			return nil, err
		}
		result.Appeals = appeals
	}

	return result, nil
}
