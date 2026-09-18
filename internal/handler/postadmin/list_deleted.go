// 查询已删除帖子列表
package postadmin

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

func ListDeletedPosts(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := postAdminService.GetDeletedPosts(page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
