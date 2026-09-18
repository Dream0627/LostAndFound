// 删除评论
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

func Delete(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 64)
		if err != nil || commentID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole, _ := middleware.CurrentRole(c)

		comment, err := commentService.GetCommentByID(commentID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		if err := commentService.CheckCommentPermission(nowUserID, nowUserRole, comment); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		if err := commentService.DeleteComment(commentID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, nil)
	}
}
