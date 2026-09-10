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
	db *repository.PostRepository
}

func NewPostService(db *repository.PostRepository) *PostService { 
	return &PostService{db: db}
}

type CreateInput struct {
    Type     string
    Title    string
    Content  string
    ImageURL *string
}

func (s *PostService) Create(input CreateInput, userID uint) (*model.Post, error) {
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
        UserID:   userID,
    }

    if err := s.db.Create(post); err != nil {
        return nil, apperror.ServerError
    }

    return post, nil
}