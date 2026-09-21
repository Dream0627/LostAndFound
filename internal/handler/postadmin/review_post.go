// 审核帖子
// 本文件对应“审核帖子”接口(把 pending 帖子改为通过/驳回)。
package postadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// ReviewPostRequest 是审核请求体：只含目标状态 status。
type ReviewPostRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReviewPost 是“审核帖子”的处理器工厂。
func ReviewPost(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request ReviewPostRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析审核请求体
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := postAdminService.ReviewPost(postID, request.Status); err != nil { // 交由业务层做状态流转校验并落库
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"post_id": postID, "status": request.Status}) // 返回审核后的帖子 id 与状态
	}
}
