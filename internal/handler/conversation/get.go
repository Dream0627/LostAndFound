// 查询单个会话详情
// 本文件对应“会话详情”接口：返回会话信息及其所属帖子快照(标题/状态/是否完成)，仅参与方可读。
// 供前端聊天页展示帖子标题、跳转原帖，并在帖子已完成时隐藏“发起完成寻找”入口。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Get 是“会话详情”的处理器工厂。
func Get(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64)
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		conversation, err := conversationService.GetConversationDetail(conversationID, nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, conversation)
	}
}
