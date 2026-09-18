package service

import (
	//"errors"
	//"fmt"
	
	"strings"
	//"time"
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

type PostService struct { 
	repository *repository.PostRepository
}

func NewPostService(repository *repository.PostRepository) *PostService { 
	return &PostService{repository: repository}
}

func (s *PostService) GetPostByID(postID uint64) (*model.Post, error) { 
	return s.repository.GetPostByID(postID)
}

func (s *PostService) GetPostByIDUnscoped(postID uint64) (*model.Post, error) { 
	return s.repository.GetPostByIDUnscoped(postID)
}

func (s *PostService) CheckpostPermission(userID uint64, role string, post *model.Post) error { 
	if role == "student" { 
		if post.UserID == userID { 
			return nil
		}
	} else if role == "postadmin" || role == "mainadmin" { 
		return nil
	}
    return apperror.UserForbiddenError
}

type CreateInput struct {
    Type     string
    Title    string
    Content  string
    ImageURL *string
}

func (s *PostService) Create(input CreateInput, userID uint64, role string) (*model.Post, error) {
	if input.Type != "lost" && input.Type != "found" {
		return nil, apperror.ParamError
	}
	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 2000 {
		return nil, apperror.ParamError
	}

	status := model.PostStatusPending
	if isPostAdmin(role) {
		status = model.PostStatusApproved
	}

	post := &model.Post{
		Type:       input.Type,
		Title:      input.Title,
		ImageURL:   input.ImageURL,
		Content:    input.Content,
		IsFinished: false,
		Status:     status,
		UserID:     userID,
	}

	if err := s.repository.Create(post); err != nil {
		return nil, apperror.ServerError
	}

	return post, nil
}

func isPostAdmin(role string) bool {
	return role == "postadmin" || role == "mainadmin"
}

type PostListResult struct {
	List     []*model.Post `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

func (s *PostService) GetPosts(types []string, statuses []string, role string, page, pageSize int) (*PostListResult, error) {
	validTypes := make([]string, 0, len(types))
	for _, postType := range types {
		if postType != "lost" && postType != "found" {
			return nil, apperror.ParamError
		}
		validTypes = append(validTypes, postType)
	}

	validStatuses := make([]string, 0, len(statuses))
	if isPostAdmin(role) {
		for _, status := range statuses {
			if status != model.PostStatusPending && status != model.PostStatusApproved && status != model.PostStatusRejected {
				return nil, apperror.InvalidPostStatusError
			}
			validStatuses = append(validStatuses, status)
		}
	} else {
		validStatuses = append(validStatuses, model.PostStatusApproved)
	}

	if len(validStatuses) == 0 {
		validStatuses = append(validStatuses, model.PostStatusPending, model.PostStatusApproved, model.PostStatusRejected)
	}

	offset := (page - 1) * pageSize
	posts, total, err := s.repository.GetPosts(validTypes, validStatuses, pageSize, offset)
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

func (s *PostService) GetVisiblePost(postID uint64, role string) (*model.Post, error) {
	post, err := s.repository.GetPostByID(postID)
	if err != nil {
		return nil, err
	}
	if !isPostAdmin(role) && post.Status != model.PostStatusApproved {
		return nil, apperror.NotFoundError
	}
	return post, nil
}

func (s *PostService) DeletePost(postID uint64) error { 
	return s.repository.DeletePost(postID)
}

func (s *PostService) RecoverPost(postID uint64) error { 
	return s.repository.RecoverPost(postID)
}