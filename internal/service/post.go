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

type CreateInput struct {
    Type     string
    Title    string
    Content  string
    ImageURL *string
}

func (s *PostService) Create(input CreateInput, userID uint64) (*model.Post, error) {
    if input.Type != "lost" && input.Type != "found" {
        return nil, apperror.ParamError
    }
    input.Content = strings.TrimSpace(input.Content)
    if len(input.Content) == 0 || len(input.Content) > 2000 {
        return nil, apperror.ParamError
    }

    post := &model.Post{
        Type:     input.Type,
        Title:    input.Title,
        ImageURL: input.ImageURL,
        Content:  input.Content,
        IsFinished: false,
        UserID:   userID,
    }

    if err := s.repository.Create(post); err != nil {
        return nil, apperror.ServerError
    }

    return post, nil
}

func (s *PostService) DeletePost(postID uint64) error { 
	return s.repository.DeletePost(postID)
}