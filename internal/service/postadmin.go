// 本文件是管理员对帖子的业务逻辑：审核、修改状态、查看已删除列表。
package service

import (
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

// PostAdminService 复用帖子仓库来读写帖子状态。
type PostAdminService struct {
	repository *repository.PostRepository
}

// NewPostAdminService 由 router 注入帖子仓库。
func NewPostAdminService(repository *repository.PostRepository) *PostAdminService {
	return &PostAdminService{repository: repository}
}

// IsValidPostStatus 判断一个状态字符串是不是三种合法取值之一。
func IsValidPostStatus(status string) bool {
	return status == model.PostStatusPending || status == model.PostStatusApproved || status == model.PostStatusRejected
}

// ReviewPost 审核帖子。规则：
//   1) 只能审核为 approved 或 rejected(不能直接改成 pending)；
//   2) 帖子必须存在；
//   3) 只有处于 pending 的帖子才能被审核，否则返回“该帖子不可审核”。
// 这样可保证“每个帖子只审核一次”，避免已处理的帖子被反复改动。
func (s *PostAdminService) ReviewPost(postID uint64, status string) error {
	if status != model.PostStatusApproved && status != model.PostStatusRejected {
		return apperror.InvalidPostStatusError
	}

	post, err := s.repository.GetPostByID(postID)
	if err != nil {
		return err
	}
	if post.Status != model.PostStatusPending {
		return apperror.PostNotPendingError
	}

	return s.repository.UpdatePostStatus(postID, status)
}

// UpdatePostStatus 直接修改帖子状态(管理员通用入口，允许改成任意合法状态)。
// 会先校验状态合法并确认帖子存在，再更新。
func (s *PostAdminService) UpdatePostStatus(postID uint64, status string) error {
	if !IsValidPostStatus(status) {
		return apperror.InvalidPostStatusError
	}

	if _, err := s.repository.GetPostByID(postID); err != nil {
		return err
	}

	return s.repository.UpdatePostStatus(postID, status)
}

// GetDeletedPosts 分页查询已删除帖子(回收站)，同样把 page/page_size 换算成 limit/offset。
func (s *PostAdminService) GetDeletedPosts(page, pageSize int) (*PostListResult, error) {
	offset := pagination.Offset(page, pageSize)
	posts, total, err := s.repository.GetDeletedPosts(pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &PostListResult{
		List:     posts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
