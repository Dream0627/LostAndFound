// 本文件对应超级管理员“反馈列表(分页)”接口，可按键 status 过滤待处理/已采纳/已驳回。
package feedback

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// List 是“反馈列表”的处理器工厂。
func List(feedbackService *service.FeedbackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status") // 审批状态过滤(可空)
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := feedbackService.List(status, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, result)
	}
}