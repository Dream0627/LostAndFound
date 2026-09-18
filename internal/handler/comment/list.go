// 查询帖子评论列表(分页)
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

func List(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		limit := 20
		offset := 0
		if v := c.Query("limit"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				limit = n
			}
		}
		if v := c.Query("offset"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n >= 0 {
				offset = n
			}
		}
		if limit > 100 {
			limit = 100
		}

		comments, err := commentService.GetCommentsByPostID(postID, limit, offset)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, comments)
	}
}
