// 查询对话消息列表
// 本文件对应“对话消息列表”接口：分页返回某对话内的消息，仅对话参与方可读。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// ListMessages 是“对话消息列表”的处理器工厂。
func ListMessages(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64) // 解析路径中的 conversation_id
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := conversationService.GetMessages(conversationID, nowUserID, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
