// 本文件对应“查询个人资料”接口(需要登录)。
package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// GetProfile 是“查询个人资料”的处理器工厂。
// 返回用户信息 + 其发布的帖子 + 其收藏的帖子(收藏夹)。
func GetProfile(userService *service.UserService, favoriteService *service.FavoriteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)

		result, err := userService.GetProfile(nowUserID) // 用令牌里的用户 ID 查资料，不接受前端指定他人 id
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		// 收藏夹属于私有数据：只回填当前登录用户本人的收藏帖子列表。
		favorites, err := favoriteService.GetFavoritesByUserID(nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		result.Favorites = favorites

		response.Success(c, result) // 返回用户信息、其帖子、其收藏夹
	}
}