// 发起申领/召领对话
// Package conversation 是“申领/召领对话 + 完成寻找”模块的 HTTP 处理器(handler)层。
// handler 只负责解析 HTTP 参数与写回响应，业务规则全部交给 service。
// 本文件对应“发起对话”接口：对某帖子发起申领(found)或召领(lost)，
// 具体语义由帖子类型决定，因此无需在接口层区分两种动作。
package conversation

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Start 是“发起申领/召领对话”的处理器工厂。
func Start(conversationService *service.ConversationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c) // 登录用户即申领/召领发起方
		nowUserRole, _ := middleware.CurrentRole(c)

		conversation, err := conversationService.StartConversation(postID, nowUserID, nowUserRole)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, conversation)
	}
}
