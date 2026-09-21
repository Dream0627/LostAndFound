// Package repository 是“数据访问层”，唯一职责就是和数据库打交道(GORM)。
// 约定：上层(service)只调用这里的方法，不直接写 SQL；
// 数据库层面的错误会在这里被翻译成领域错误(哨兵错误)，向上抛出，
// 使业务层不用关心底层是 GORM 还是 MySQL。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
)

// ErrCommentNotFound 是本层定义的“哨兵错误”，表示“评论不存在”。
// 用 errors.New 定义成一个固定值，上层就能用 errors.Is 精确判断这种错误。
var (
	ErrCommentNotFound = errors.New("comment not found")
)

// CommentRepository 持有数据库句柄 db，为评论提供增删查能力。
type CommentRepository struct {
	db *gorm.DB
}

// NewCommentRepository 是构造函数：由 router 在装配时注入 db。
func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{db: db}
}

// Create 插入一条评论。传入的是指针，GORM 会把自增主键回填到该结构体上。
func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

// GetCommentByID 按主键查询一条(未删除的)评论。
// 若查不到，GORM 返回 gorm.ErrRecordNotFound，这里把它翻译成 ErrCommentNotFound。
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

// GetCommentsByPostID 分页查询某帖子的评论，并返回总数(total 用于前端算总页数)。
// buildQuery 用一个闭包复用了“基础条件(post_id)”，避免写两遍导致两处条件不一致。
// 查询两次：一次 Count 求总数，一次带 Limit/Offset 取当页数据。
// 参数名 limit/offset 是数据库层词汇；service 层会换算并对外暴露 page/page_size。
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

// DeleteComment 按主键删除评论。因模型含 gorm.DeletedAt 字段，
// 这里实际执行的是“软删除”：只更新 deleted_at，数据仍保留在库中。
func (r *CommentRepository) DeleteComment(commentID uint64) error {
	return r.db.Delete(&model.Comment{}, commentID).Error
}
