// 删除帖子
package post

import (
	"strconv"
	"errors"
	//"fmt"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

func DeletePost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole , _ := middleware.CurrentRole(c)
		post, err := postService.GetPostByID(postID)
		
		if err != nil {
			if errors.Is(err, apperror.NotFoundError) {
				apperror.AbortWithException(c, apperror.NotFoundError, nil)
			} else {
				apperror.AbortWithError(c, err)
			}
			return
		}

		if err := postService.CheckpostPermission(nowUserID, nowUserRole, post); err != nil {
			apperror.AbortWithException(c, apperror.UserForbiddenError, nil)
			return
		}

		if err := postService.DeletePost(postID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, gin.H{"post_id": postID})
	}
}