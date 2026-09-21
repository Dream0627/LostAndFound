// 查询帖子列表
// 本文件对应“查询帖子列表”接口。支持按类型、状态过滤并分页。
package post

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// ListPosts 是“帖子列表”的处理器工厂。
func ListPosts(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		types := c.QueryArray("type") // 读取可重复的 type 查询参数(?type=lost&type=found)
		statuses := c.QueryArray("status") // 读取可重复的 status 查询参数
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		nowUserRole, _ := middleware.CurrentRole(c)

		result, err := postService.GetPosts(types, statuses, nowUserRole, page, pageSize) // 交由业务层做过滤与可见性控制
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
