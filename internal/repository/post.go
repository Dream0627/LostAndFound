package repository

import (
	//"errors"
	//"fmt"

	//"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	//"LAF/pkg/apperror"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *model.Post) error { 
	return r.db.Create(post).Error
}