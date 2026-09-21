// 查询帖子评论列表(分页)
// 本文件对应“查询某帖子的评论列表(分页)”接口。该接口公开，无需登录。
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// List 是“评论列表”的处理器工厂。
func List(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size")) // 解析分页参数(统一默认值/上限)

		comments, err := commentService.GetCommentsByPostID(postID, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, comments) // 返回 {list,total,page,page_size}
	}
}
