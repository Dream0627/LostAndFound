// 查询帖子详情
// 本文件对应“查询帖子详情”接口。该接口用可选鉴权，未登录也能看公开帖子。
package post

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// GetPost 是“帖子详情”的处理器工厂。
func GetPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserRole, _ := middleware.CurrentRole(c)

		post, err := postService.GetVisiblePost(postID, nowUserRole) // 按角色返回“可见的”帖子(未审核帖子对普通用户不可见)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, post) // 返回帖子详情
	}
}
