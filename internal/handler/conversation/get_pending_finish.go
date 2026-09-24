// 查询待处理的完成寻找申请
// 本文件对应“查询会话当前待处理的完成寻找申请”接口：仅参与方可见，供前端进入会话时判断是否显示申请横幅。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// GetPendingFinish 是“查询待处理的完成寻找申请”的处理器工厂；无待处理申请时 data 为 null。
func GetPendingFinish(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		conversationID, err := strconv.ParseUint(c.Param("conversation_id"), 10, 64) // 解析路径中的 conversation_id
		if err != nil || conversationID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)

		request, err := conversationService.GetPendingFinishRequest(conversationID, nowUserID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, request) // 无待处理申请时为 nil
	}
}