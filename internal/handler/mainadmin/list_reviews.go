// 本文件对应超级管理员“待审批列表”接口(查看帖子发布 / 注销申诉的待审批请求)。
package mainadmin

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// ListReviews 是“待审批列表”的处理器工厂。
// 查询参数 type 可选：post(仅帖子发布) / appeal(仅注销申诉)；不传则两类都返回。
func ListReviews(mainAdminService *service.MainAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		reviewType := c.Query("type") // 待审批类型过滤(可空)
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := mainAdminService.GetPendingReviews(reviewType, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
