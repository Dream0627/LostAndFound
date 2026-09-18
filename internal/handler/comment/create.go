// 发表评论
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type CreateCommentRequest struct {
	Content string `json:"content" binding:"required"`
}

func Create(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		var request CreateCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdComment, err := commentService.Create(postID, nowUserID, service.CreateCommentInput{
			Content: request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdComment)
	}
}
