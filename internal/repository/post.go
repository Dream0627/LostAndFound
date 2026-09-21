// 本文件是帖子的数据访问层，包含增删查改以及“软删除/恢复”“审核状态更新”等操作。
// 同样遵循：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	"errors"
	//"fmt"

	//"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// PostRepository 持有数据库句柄 db，为帖子提供数据访问能力。
type PostRepository struct {
	db *gorm.DB
}


// NewPostRepository 是构造函数，由 router 注入 db。
func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

// GetPostByID 查询未删除的帖子。
// 查不到时返回 apperror.PostNotFoundError；其他数据库异常统一返回 apperror.DatabaseError。
func (r *PostRepository) GetPostByID(postID uint64) (*model.Post, error) { 
	var post model.Post
	err := r.db.Where("id = ?", postID).First(&post).Error
	if err != nil { 
		if errors.Is(err, gorm.ErrRecordNotFound) { 
			return nil, apperror.PostNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &post, nil
}

// GetPostByIDUnscoped 查询“包含已软删除在内”的帖子。
// Unscoped() 会跳过 GORM 默认的“自动过滤 deleted_at IS NULL”，
// 因此能查到已删除的帖子，用于“恢复帖子”等场景。
func (r *PostRepository) GetPostByIDUnscoped(postID uint64) (*model.Post, error) { 
	var post model.Post
	err := r.db.Unscoped().Where("id = ?", postID).First(&post).Error
	if err != nil { 
		if errors.Is(err, gorm.ErrRecordNotFound) { 
			return nil, apperror.PostNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &post, nil
}

// Create 插入一条帖子，自增主键会回填到 post.ID。
func (r *PostRepository) Create(post *model.Post) error { 
	if err := r.db.Create(post).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// DeletePost 删除帖子，并在同一事务里一并软删除其下所有评论。
// 为什么要手动删评论？因为软删除只是 UPDATE，不会触发数据库外键级联；
// 用 Transaction 保证“删评论 + 删帖子”要么都成功、要么都回滚，避免出现半删状态。
func (r *PostRepository) DeletePost(postID uint64) error { 
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", postID).Delete(&model.Comment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Post{}, postID).Error
	})
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// RecoverPost 恢复被软删除的帖子：把 deleted_at 重新置为 NULL。
// 必须用 Unscoped() 才能操作到已软删除的记录(否则 GORM 会自动加上“未删除”条件)。
func (r *PostRepository) RecoverPost(postID uint64) error { 
	if err := r.db.Unscoped().Model(&model.Post{}).Where("id = ?", postID).Update("deleted_at", nil).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetPosts 分页查询帖子，支持按 type / status 过滤。
// 过滤条件用“仅当有值时才拼接”的方式，实现对空过滤条件的忽略；
// 同样用闭包复用基础查询，先 Count 求总数，再 Limit/Offset 取当页数据，按 id 倒序(新帖在前)。
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
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return posts, total, nil
}

// GetDeletedPosts 分页查询“已删除”的帖子(管理员回收站)。
// 通过 Unscoped() + deleted_at IS NOT NULL 精确筛出被软删除的记录。
func (r *PostRepository) GetDeletedPosts(limit, offset int) ([]*model.Post, int64, error) {
	var posts []*model.Post
	var total int64

	buildQuery := func() *gorm.DB {
		return r.db.Unscoped().Model(&model.Post{}).Where("deleted_at IS NOT NULL")
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return posts, total, nil
}

// UpdatePostStatus 更新帖子的审核状态(如 pending -> approved / rejected)。
func (r *PostRepository) UpdatePostStatus(postID uint64, status string) error {
	if err := r.db.Model(&model.Post{}).Where("id = ?", postID).Update("status", status).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}
