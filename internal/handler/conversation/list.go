// 查询我的对话列表
// 本文件对应“我的对话列表”接口：返回当前用户作为发起方或楼主参与的全部对话。
package conversation

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// List 是“我的对话列表”的处理器工厂。
func List(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := conversationService.GetMyConversations(nowUserID, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
