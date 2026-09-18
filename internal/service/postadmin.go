package service

import (
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

type PostAdminService struct {
	repository *repository.PostRepository
}

func NewPostAdminService(repository *repository.PostRepository) *PostAdminService {
	return &PostAdminService{repository: repository}
}

func IsValidPostStatus(status string) bool {
	return status == model.PostStatusPending || status == model.PostStatusApproved || status == model.PostStatusRejected
}

// ReviewPost 管理员审核帖子,仅允许审核中(pending)的帖子通过或驳回,不可再次审核
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

// UpdatePostStatus 管理员直接修改帖子状态(任意状态)
func (s *PostAdminService) UpdatePostStatus(postID uint64, status string) error {
	if !IsValidPostStatus(status) {
		return apperror.InvalidPostStatusError
	}

	if _, err := s.repository.GetPostByID(postID); err != nil {
		return err
	}

	return s.repository.UpdatePostStatus(postID, status)
}

func (s *PostAdminService) GetDeletedPosts(page, pageSize int) (*PostListResult, error) {
	offset := (page - 1) * pageSize
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
