// Package feedback 是反馈模块的 HTTP 处理器层。
// 与其它 handler 一致：只做“解析请求参数、写回响应”，业务规则交给 service。
// 本文件对应“提交反馈”接口(登录用户提交)。
package feedback

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// SubmitRequest 是提交反馈的请求体。提交人(user_id)取自登录令牌，不由请求体传入。
type SubmitRequest struct {
	Content string `json:"content" binding:"required"`
}

// Submit 是“提交反馈”的处理器工厂。
func Submit(feedbackService *service.FeedbackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := middleware.CurrentUserID(c) // 从令牌解析出的当前用户 ID(提交人)
		if !ok {
			apperror.AbortWithError(c, apperror.UnauthorizedError)
			return
		}

		var request SubmitRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		created, err := feedbackService.Submit(userID, service.FeedbackInput{Content: request.Content})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, created)
	}
}