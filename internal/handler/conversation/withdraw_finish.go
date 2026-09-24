// 撤回完成寻找申请
// 本文件对应“撤回完成寻找申请”接口：由申请人(发起方)撤回自己发起的待处理申请，帖子不受影响。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// WithdrawFinish 是“撤回完成寻找申请”的处理器工厂。
func WithdrawFinish(conversationService *service.ConversationService) gin.HandlerFunc {
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

		nowUserID, _ := middleware.CurrentUserID(c)

		if err := conversationService.WithdrawFinishRequest(conversationID, requestID, nowUserID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, nil)
	}
}