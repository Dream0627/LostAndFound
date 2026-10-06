package feedback

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

type FeedbackCreateRequest struct {
	FeedbackID uint64
	Content    string
}

func Create(feedbackService *service.FeedbackService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c) // 从令牌解析出的当前用户 ID(作者)，忽略第二个返回值(是否取到)

		var request FeedbackCreateRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdFeedback, err := feedbackService.CreateFeedback(service.FeedbackCreateInput{ // 调用业务层创建反馈，作者以登录身份为准
			UserID:  nowUserID,
			Content: request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdFeedback) // 成功：返回创建好的反馈
	}
}
