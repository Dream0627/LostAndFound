// 本文件对应“恢复帖子”接口：把被软删除的帖子恢复为正常状态。
package post

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// RecoverPost 是“恢复帖子”的处理器工厂。
func RecoverPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole, _ := middleware.CurrentRole(c)
		post, err := postService.GetPostByIDUnscoped(postID) // 注意：必须用 Unscoped 版本才能查到已删除的帖子

		if err != nil {
			if errors.Is(err, apperror.PostNotFoundError) {
				apperror.AbortWithException(c, apperror.PostNotFoundError, nil)
			} else {
				apperror.AbortWithError(c, err)
			}
			return
		}

		if err := postService.CheckpostPermission(nowUserID, nowUserRole, post); err != nil { // 权限校验：本人或管理员才可恢复
			apperror.AbortWithException(c, apperror.UserForbiddenError, nil)
			return
		}

		if err := postService.RecoverPost(postID); err != nil { // 执行恢复(deleted_at 置回 NULL)
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, gin.H{"post_id": postID})
	}

}
