// 发送对话消息
// 本文件对应“发送消息”接口：在当前对话中追加一条消息，仅对话参与方可发。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// SendMessageRequest 是发送消息的请求体结构。content 必传。
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// SendMessage 是“发送消息”的处理器工厂。
func SendMessage(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64) // 解析路径中的 conversation_id
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request SendMessageRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		message, err := conversationService.SendMessage(conversationID, nowUserID, request.Content)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, message)
	}
}
