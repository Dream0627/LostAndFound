// 查询帖子公告列表(分页)
// 本文件对应“查询某帖子的公告列表(分页)”接口。该接口公开，无需登录。
package announcement

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// List 是“公告列表”的处理器工厂。
func List(announcementService *service.AnnouncementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size")) // 解析分页参数(统一默认值/上限)
		announcements, err := announcementService.GetAnnouncements(page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, announcements) // 返回 {list,total,page,page_size}
	}
}
