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

		// finished 解析“已完成/未完成”筛选项：空=不限，true=已完成，false=未完成；非法值按参数错误处理。
		finished, err := parseFinished(c.Query("finished"))
		if err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserRole, _ := middleware.CurrentRole(c)

		result, err := postService.GetPosts(types, statuses, finished, nowUserRole, page, pageSize) // 交由业务层做过滤与可见性控制
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, result)
	}
}

// parseFinished 解析 finished 查询参数(已完成/未完成筛选项)。
// 取值：空或缺省表示不限(返回 nil)；"true" 只看已完成；"false" 只看未完成；其他值按参数错误处理。
func parseFinished(raw string) (*bool, error) {
	switch raw {
	case "":
		return nil, nil
	case "true":
		value := true
		return &value, nil
	case "false":
		value := false
		return &value, nil
	default:
		return nil, apperror.ParamError
	}
}
