// Package appeal 是申诉模块的 HTTP 处理器层。
// 与其它 handler 一致：只做“解析请求参数、写回响应”，业务规则交给 service。
// 本文件对应“提交申诉”接口(公开接口，无需登录)。
package appeal

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// CreateAppealRequest 是提交申诉的请求体。
// username 指明被注销的账号；reason 为申诉原因；content 为申诉说明(可选，reason=other 时必填)。
type CreateAppealRequest struct {
	Username string `json:"username" binding:"required"`
	Reason   string `json:"reason" binding:"required"`
	Content  string `json:"content"`
}

// Create 是“提交申诉”的处理器工厂。
func Create(appealService *service.AppealService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request CreateAppealRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体(必填字段缺失则报错)
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdAppeal, err := appealService.Create(service.CreateAppealInput{
			Username: request.Username,
			Reason:   request.Reason,
			Content:  request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdAppeal) // 成功：返回创建好的申诉
	}
}
