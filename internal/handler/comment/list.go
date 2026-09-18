// 查询帖子评论列表(分页)
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

func List(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		comments, err := commentService.GetCommentsByPostID(postID, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, comments)
	}
}
