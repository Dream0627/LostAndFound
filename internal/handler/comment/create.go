// 发表评论
// Package comment 是评论模块的 HTTP 处理器(handler)层。
// handler 的职责很单一：只做“解析 HTTP 请求参数”和“写回 HTTP 响应”，
// 真正的业务规则交给 service，绝不在这里直接操作数据库。
// 本文件对应“发表评论”接口。
package comment

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// CreateCommentRequest 是发表评论的请求体结构。
// json 标签：声明请求体里的字段名(post_id / content)。
// binding:"required"：Gin 在解析时会校验这两个字段必传，缺失则解析报错。
type CreateCommentRequest struct {
	PostID  uint64 `json:"post_id" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// Create 是“发表评论”的处理器工厂。
// 之所以返回 gin.HandlerFunc 而不是直接写处理逻辑，是为了在装配时把 service 依赖“闭包捕获”进来，
// 这样路由注册处只需传入依赖即可，依赖注入更清晰。
func Create(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c) // 从令牌解析出的当前用户 ID(作者)，忽略第二个返回值(是否取到)

		var request CreateCommentRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdComment, err := commentService.Create(nowUserID, service.CreateCommentInput{ // 调用业务层创建评论，作者以登录身份为准
			PostID:  request.PostID,
			Content: request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdComment) // 成功：返回创建好的评论
	}
}
