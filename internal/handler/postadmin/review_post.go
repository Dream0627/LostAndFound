// 审核帖子
package postadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type ReviewPostRequest struct {
	Status string `json:"status" binding:"required"`
}

func ReviewPost(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request ReviewPostRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := postAdminService.ReviewPost(postID, request.Status); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"post_id": postID, "status": request.Status})
	}
}
