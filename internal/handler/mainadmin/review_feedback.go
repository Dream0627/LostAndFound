// 本文件对应超级管理员“审批反馈”接口(把 pending 反馈改为通过/驳回)。
package mainadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// ReviewFeedbackRequest 是审批反馈的请求体：只含目标状态 status。
type ReviewFeedbackRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReviewFeedback 是“审批反馈”的处理器工厂。
func ReviewFeedback(feedbackService *service.FeedbackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		feedbackID, err := strconv.ParseUint(c.Param("feedback_id"), 10, 64) // 解析路径中的 feedback_id
		if err != nil || feedbackID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request ReviewFeedbackRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := feedbackService.Review(feedbackID, request.Status); err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, gin.H{"feedback_id": feedbackID, "status": request.Status})
	}
}