// 查询帖子列表
package post

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

func ListPosts(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		types := c.QueryArray("type")
		statuses := c.QueryArray("status")
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		nowUserRole, _ := middleware.CurrentRole(c)

		result, err := postService.GetPosts(types, statuses, nowUserRole, page, pageSize)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}
