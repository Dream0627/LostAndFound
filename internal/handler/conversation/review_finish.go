// 处理完成寻找申请
// 本文件对应“处理完成寻找申请”接口：由发起方之外的另一方同意(agreed)或拒绝(rejected)；
// 同意则把对应帖子置为已完成。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// ReviewFinishRequest 是处理完成申请的请求体结构。status 必传(agreed/rejected)。
type ReviewFinishRequest struct {
	Status string `json:"status" binding:"required"`
}

// ReviewFinish 是“处理完成寻找申请”的处理器工厂。
func ReviewFinish(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64) // 解析路径中的 conversation_id
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}
		requestID, err := strconv.ParseUint(c.Param("request_id"), 10, 64) // 解析路径中的 request_id
		if err != nil || requestID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request ReviewFinishRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		result, err := conversationService.ReviewFinishRequest(conversationID, requestID, nowUserID, request.Status)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
