// 发表评论
package comment

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type CreateCommentRequest struct {
	PostID  uint64 `json:"post_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func Create(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)

		var request CreateCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdComment, err := commentService.Create(nowUserID, service.CreateCommentInput{
			PostID:  request.PostID,
			Content: request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdComment)
	}
}
