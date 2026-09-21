// 修改帖子审核状态
// 本文件对应“修改帖子状态”接口(管理员通用入口，可改为任意合法状态)。
package postadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// UpdatePostStatusRequest 是改状态请求体：只含目标状态 status。
type UpdatePostStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// UpdatePostStatus 是“修改帖子状态”的处理器工厂。
func UpdatePostStatus(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request UpdatePostStatusRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := postAdminService.UpdatePostStatus(postID, request.Status); err != nil { // 校验状态合法+帖子存在后更新
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"post_id": postID, "status": request.Status}) // 返回帖子 id 与更新后的状态
	}
}
