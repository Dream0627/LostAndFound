package repository

import (
	"errors"
	//"fmt"

	//"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

type PostRepository struct {
	db *gorm.DB
}


func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) GetPostByID(postID uint64) (*model.Post, error) { 
	var post model.Post
	err := r.db.Where("id = ?", postID).First(&post).Error
	if err != nil { 
		if errors.Is(err, gorm.ErrRecordNotFound) { 
			return nil, apperror.NotFoundError
		}
		return nil, apperror.ServerError
	}
	return &post, nil
}

func (r *PostRepository) GetPostByIDUnscoped(postID uint64) (*model.Post, error) { 
	var post model.Post
	err := r.db.Unscoped().Where("id = ?", postID).First(&post).Error
	if err != nil { 
		if errors.Is(err, gorm.ErrRecordNotFound) { 
			return nil, apperror.NotFoundError
		}
		return nil, apperror.ServerError
	}
	return &post, nil
}

func (r *PostRepository) Create(post *model.Post) error { 
	return r.db.Create(post).Error
}

func (r *PostRepository) DeletePost(postID uint64) error { 
	return r.db.Delete(&model.Post{}, postID).Error
}

func (r *PostRepository) RecoverPost(postID uint64) error { 
	return r.db.Unscoped().Model(&model.Post{}).Where("id = ?", postID).Update("deleted_at", nil).Error
}
