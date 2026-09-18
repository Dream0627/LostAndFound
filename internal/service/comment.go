package service

import (
	"errors"
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

type CommentService struct {
	repository     *repository.CommentRepository
	postRepository *repository.PostRepository
}

func NewCommentService(repository *repository.CommentRepository, postRepository *repository.PostRepository) *CommentService {
	return &CommentService{
		repository:     repository,
		postRepository: postRepository,
	}
}

type CreateCommentInput struct {
	Content string `json:"content"`
}

func (s *CommentService) Create(postID uint64, userID uint64, input CreateCommentInput) (*model.Comment, error) {
	if _, err := s.postRepository.GetPostByID(postID); err != nil {
		return nil, err
	}

	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 1000 {
		return nil, apperror.ParamError
	}

	comment := &model.Comment{
		PostID:  postID,
		UserID:  userID,
		Content: input.Content,
	}

	if err := s.repository.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) GetCommentByID(commentID uint64) (*model.Comment, error) {
	comment, err := s.repository.GetCommentByID(commentID)
	if errors.Is(err, repository.ErrCommentNotFound) {
		return nil, apperror.CommentNotFoundError
	}
	if err != nil {
		return nil, err
	}
	return comment, nil
}

type CommentListResult struct {
	List     []*model.Comment `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

func (s *CommentService) GetCommentsByPostID(postID uint64, page, pageSize int) (*CommentListResult, error) {
	offset := (page - 1) * pageSize
	comments, total, err := s.repository.GetCommentsByPostID(postID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &CommentListResult{
		List:     comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *CommentService) CheckCommentPermission(userID uint64, role string, comment *model.Comment) error {
	if role == "student" {
		if comment.UserID == userID {
			return nil
		}
	} else if role == "postadmin" || role == "mainadmin" {
		return nil
	}
	return apperror.UserForbiddenError
}

func (s *CommentService) DeleteComment(commentID uint64) error {
	return s.repository.DeleteComment(commentID)
}
