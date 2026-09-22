// 本文件对应超级管理员“审核申诉”接口(把 pending 申诉改为通过/驳回)。
package mainadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// ReviewAppealRequest 是审核申诉的请求体：只含目标状态 status。
type ReviewAppealRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReviewAppeal 是“审核申诉”的处理器工厂。
func ReviewAppeal(mainAdminService *service.MainAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		appealID, err := strconv.ParseUint(c.Param("appeal_id"), 10, 64) // 解析路径中的 appeal_id
		if err != nil || appealID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request ReviewAppealRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析审核请求体
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := mainAdminService.ReviewAppeal(appealID, request.Status); err != nil { // 交给业务层做状态流转与级联恢复
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"appeal_id": appealID, "status": request.Status})
	}
}
