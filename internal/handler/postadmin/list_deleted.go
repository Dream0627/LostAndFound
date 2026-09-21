// 查询已删除帖子列表
// Package postadmin 是帖子管理员相关的 HTTP 处理器层。
// 本文件对应“查询已删除帖子列表(回收站)”接口，路由上已用 RequireRole 限定管理员。
package postadmin

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
	"LAF/pkg/response"
)

// ListDeletedPosts 是“已删除帖子列表”的处理器工厂。
func ListDeletedPosts(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := pagination.Parse(c.Query("page"), c.Query("page_size"))

		result, err := postAdminService.GetDeletedPosts(page, pageSize) // 分页查询软删除的帖子
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result) // 返回列表信封
	}
}
