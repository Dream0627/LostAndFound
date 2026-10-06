// 本文件对应“取消收藏”接口。收藏人取自登录令牌，帖子 ID 取自路径。
package favorite

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Remove 是“取消收藏”的处理器工厂(幂等)。
func Remove(favoriteService *service.FavoriteService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.CurrentUserID(c) // 取消人=当前登录用户
		if !ok {
			apperror.AbortWithError(c, apperror.UnauthorizedError)
			return
		}

		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		if err := favoriteService.Remove(userID, postID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, gin.H{"post_id": postID, "favorited": false})
	}
}