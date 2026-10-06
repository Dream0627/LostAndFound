// 本文件是收藏帖子的数据访问层：创建、查询“活跃收藏”、取消(软删)、按用户查收藏的帖子列表。
// 与其它 repository 一致：只和数据库打交道，并把数据库错误翻译成领域错误(apperror)向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// ErrFavoriteNotFound 哨兵错误：未找到“活跃收藏”(该用户当前未收藏该帖子)。
var ErrFavoriteNotFound = errors.New("favorite not found")

// FavoriteRepository 持有数据库句柄 db，为收藏帖子提供数据访问能力。
type FavoriteRepository struct {
	db *gorm.DB
}

// NewFavoriteRepository 是构造函数，由 router 在装配时注入 db。
func NewFavoriteRepository(db *gorm.DB) *FavoriteRepository {
	return &FavoriteRepository{db: db}
}

// Create 插入一条收藏，自增主键会回填到 favorite.ID。
func (r *FavoriteRepository) Create(favorite *model.Favorite) error {
	if err := r.db.Create(favorite).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetActive 查询“未删除(活跃)”的收藏记录；查不到时返回 ErrFavoriteNotFound。
// GORM 默认会过滤 deleted_at 非空的行，正好得到“当前仍在收藏”的记录。
func (r *FavoriteRepository) GetActive(userID, postID uint64) (*model.Favorite, error) {
	var favorite model.Favorite
	err := r.db.Where("user_id = ? AND post_id = ?", userID, postID).First(&favorite).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFavoriteNotFound
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &favorite, nil
}

// Delete 软删除一条收藏(@favorite 需带主键 ID)，对应“取消收藏”。
func (r *FavoriteRepository) Delete(favorite *model.Favorite) error {
	if err := r.db.Delete(favorite).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetFavoritePosts 查询某用户收藏的帖子列表(按收藏时间倒序)，并回填作者姓名。
// 用 JOIN 到 favorites 表：只取“当前仍活跃收藏(favorites.deleted_at IS NULL)”的帖子，
// 帖子本身的软删除过滤(posts.deleted_at IS NULL)由 GORM 依据 model.Post 自动加上。
func (r *FavoriteRepository) GetFavoritePosts(userID uint64) ([]*model.Post, error) {
	var posts []*model.Post
	err := r.db.Model(&model.Post{}).
		Select("posts.*").
		Joins("JOIN favorites ON favorites.post_id = posts.id AND favorites.deleted_at IS NULL").
		Where("favorites.user_id = ?", userID).
		Order("favorites.id desc").
		Find(&posts).Error
	if err != nil {
		return nil, apperror.DatabaseError
	}
	r.fillPostAuthorNames(posts)
	return posts, nil
}

// fillPostAuthorNames 为一批帖子批量填充作者姓名(AuthorName)。
func (r *FavoriteRepository) fillPostAuthorNames(posts []*model.Post) {
	if len(posts) == 0 {
		return
	}
	ids := make([]uint64, 0, len(posts))
	for _, p := range posts {
		if p != nil {
			ids = append(ids, p.UserID)
		}
	}
	names := loadUserNames(r.db, collectUserIDs(ids))
	for _, p := range posts {
		if p == nil {
			continue
		}
		if name, ok := names[p.UserID]; ok && name != "" {
			p.AuthorName = name
		} else {
			p.AuthorName = authorNameUnknown
		}
	}
}