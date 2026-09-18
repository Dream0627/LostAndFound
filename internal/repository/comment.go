package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) GetCommentByID(commentID uint64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Where("id = ?", commentID).First(&comment).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCommentNotFound
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) GetCommentsByPostID(postID uint64, limit, offset int) ([]*model.Comment, int64, error) {
	var comments []*model.Comment
	var total int64

	buildQuery := func() *gorm.DB {
		return r.db.Model(&model.Comment{}).Where("post_id = ?", postID)
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}
	return comments, total, nil
}

func (r *CommentRepository) DeleteComment(commentID uint64) error {
	return r.db.Delete(&model.Comment{}, commentID).Error
}
