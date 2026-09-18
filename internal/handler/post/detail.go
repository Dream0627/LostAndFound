// 查询帖子详情
package post

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

func GetPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserRole, _ := middleware.CurrentRole(c)

		post, err := postService.GetVisiblePost(postID, nowUserRole)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, post)
	}
}
