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
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Post{}, postID).Error
	})
}

func (r *PostRepository) RecoverPost(postID uint64) error { 
	return r.db.Unscoped().Model(&model.Post{}).Where("id = ?", postID).Update("deleted_at", nil).Error
}

func (r *PostRepository) GetPosts(types []string, statuses []string, limit, offset int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	buildQuery := func() *gorm.DB {
		query := r.db.Model(&model.Post{})
		if len(types) > 0 {
			query = query.Where("type IN ?", types)
		}
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
		return query
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.ServerError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, apperror.ServerError
	}
	return posts, total, nil
}

func (r *PostRepository) GetDeletedPosts(limit, offset int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	buildQuery := func() *gorm.DB {
		return r.db.Unscoped().Model(&model.Post{}).Where("deleted_at IS NOT NULL")
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.ServerError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, apperror.ServerError
	}
	return posts, total, nil
}

func (r *PostRepository) UpdatePostStatus(postID uint64, status string) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Update("status", status).Error
}
