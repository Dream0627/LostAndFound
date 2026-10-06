package postadmin

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// CreateAnnouncementRequest 是发表公告的请求体结构。
// json 标签：声明请求体里的字段名(title / content)。
// binding:"required"：Gin 在解析时会校验这两个字段必传，缺失则解析报错。
// 发布者(admin_id)取自登录令牌，不由请求体传入。
type CreateAnnouncementRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// CreateAnnouncement 是“发表公告”的处理器工厂。
// 之所以返回 gin.HandlerFunc 而不是直接写处理逻辑，是为了在装配时把 service 依赖“闭包捕获”进来，
// 这样路由注册处只需传入依赖即可，依赖注入更清晰。
func CreateAnnouncement(announcementService *service.AnnouncementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowAdminID, ok := middleware.CurrentUserID(c) // 从令牌解析出的当前用户 ID(作者)
		if !ok {
			apperror.AbortWithError(c, apperror.UnauthorizedError)
			return
		}
		var request CreateAnnouncementRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析并校验请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		createdAnnouncement, err := announcementService.Create(nowAdminID, service.AnnouncementInput{ // 调用业务层创建公告，作者以登录身份为准
			Title:   request.Title,
			Content: request.Content,
		})
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdAnnouncement) // 成功：返回创建好的公告
	}
}
