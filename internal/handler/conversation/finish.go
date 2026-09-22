// 发起完成寻找申请
// 本文件对应“发起完成寻找申请”接口：对话任一方均可发起，等待另一方同意/拒绝。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Finish 是“发起完成寻找申请”的处理器工厂。
func Finish(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64) // 解析路径中的 conversation_id
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		request, err := conversationService.CreateFinishRequest(conversationID, nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, request)
	}
}
