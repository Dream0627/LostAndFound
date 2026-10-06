// 本文件是收藏帖子的业务逻辑层：收藏、取消收藏、查询本人收藏列表。
package service

import (
	"errors"

	"LAF/internal/model"
	"LAF/internal/repository"
)

// FavoriteService 依赖收藏仓库和帖子仓库：收藏/取消收藏需校验帖子存在，
// 查收藏列表则联表返回帖子并回填作者姓名。
type FavoriteService struct {
	repository     *repository.FavoriteRepository
	postRepository *repository.PostRepository
}

// NewFavoriteService 由 router 在装配阶段调用，注入收藏仓库与帖子仓库。
func NewFavoriteService(repository *repository.FavoriteRepository, postRepository *repository.PostRepository) *FavoriteService {
	return &FavoriteService{repository: repository, postRepository: postRepository}
}

// Add 收藏帖子(幂等)。规则：
//  1. 帖子必须存在(且未被删除)，否则返回帖子不存在的错误；
//  2. 若已收藏(存在活跃收藏)，直接返回成功，不重复插入。
func (s *FavoriteService) Add(userID, postID uint64) error {
	if _, err := s.postRepository.GetPostByID(postID); err != nil {
		return err
	}

	if _, err := s.repository.GetActive(userID, postID); err == nil {
		return nil // 已收藏，幂等返回
	} else if !errors.Is(err, repository.ErrFavoriteNotFound) {
		return err
	}

	return s.repository.Create(&model.Favorite{UserID: userID, PostID: postID})
}

// Remove 取消收藏(幂等)。规则：若当前未收藏，直接返回成功(不报错)；已收藏则软删除。
func (s *FavoriteService) Remove(userID, postID uint64) error {
	favorite, err := s.repository.GetActive(userID, postID)
	if errors.Is(err, repository.ErrFavoriteNotFound) {
		return nil // 本来就没收藏，幂等返回
	}
	if err != nil {
		return err
	}
	return s.repository.Delete(favorite)
}

// GetFavoritesByUserID 查询某用户的收藏帖子列表(供个人档案“收藏夹”展示)。
func (s *FavoriteService) GetFavoritesByUserID(userID uint64) ([]*model.Post, error) {
	return s.repository.GetFavoritePosts(userID)
}