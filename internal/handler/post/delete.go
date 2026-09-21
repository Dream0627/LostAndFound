// 删除帖子
// 本文件对应“删除帖子”接口(软删除)。
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

// DeletePost 是“删除帖子”的处理器工厂。
func DeletePost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole , _ := middleware.CurrentRole(c)
		post, err := postService.GetPostByID(postID) // 先查帖子，用于权限判断+确认存在
		
		if err != nil {
			if errors.Is(err, apperror.NotFoundError) {
				apperror.AbortWithException(c, apperror.NotFoundError, nil)
			} else {
				apperror.AbortWithError(c, err)
			}
			return
		}

		if err := postService.CheckpostPermission(nowUserID, nowUserRole, post); err != nil { // 权限校验：本人或管理员才可删
			apperror.AbortWithException(c, apperror.UserForbiddenError, nil)
			return
		}

		if err := postService.DeletePost(postID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, gin.H{"post_id": postID}) // 返回被删除帖子的 id
	}
}